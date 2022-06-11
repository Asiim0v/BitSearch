package searcher

import (
	"BitSearch/searcher/arrays"
	"BitSearch/searcher/model"
	"BitSearch/searcher/pagination"
	"BitSearch/searcher/sorts"
	"BitSearch/searcher/statistic"
	"BitSearch/searcher/storage"
	"BitSearch/searcher/utils"
	"BitSearch/searcher/words"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Engine struct {
	IndexPath string  //索引文件存储目录
	Option    *Option //配置

	invertedIndexStorages []*storage.LeveldbStorage //关键字和Id映射，倒排索引,key=id,value=[]words
	positiveIndexStorages []*storage.LeveldbStorage //ID和key映射，用于计算相关度，一个id 对应多个key，正排索引
	docStorages           []*storage.LeveldbStorage //文档仓

	sync.Mutex                                   //锁
	sync.WaitGroup                               //等待
	addDocumentWorkerChan []chan *model.IndexDoc //添加索引的通道
	IsDebug               bool                   //是否调试模式
	Tokenizer             *words.Tokenizer       //分词器
	DatabaseName          string                 //数据库名

	Shard   int   //分片数
	Timeout int64 //超时时间,单位秒

	Recorder *statistic.Trie //搜索统计
}

type Option struct {
	InvertedIndexName string //倒排索引
	PositiveIndexName string //正排索引
	DocIndexName      string //文档存储
}

// Init 初始化索引引擎
func (e *Engine) Init() {
	e.Add(1)
	defer e.Done()

	if e.Option == nil {
		e.Option = e.GetOptions()
	}
	if e.Timeout == 0 {
		e.Timeout = 10 * 60 //默认10分钟
	}
	//log.Println("数据存储目录：", e.IndexPath)

	e.addDocumentWorkerChan = make([]chan *model.IndexDoc, e.Shard)
	//初始化文件存储
	for shard := 0; shard < e.Shard; shard++ {

		//初始化chan
		worker := make(chan *model.IndexDoc, 1000)
		e.addDocumentWorkerChan[shard] = worker

		//初始化chan
		go e.DocumentWorkerExec(worker)

		s, err := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.DocIndexName, shard)), e.Timeout)
		if err != nil {
			panic(err)
		}
		e.docStorages = append(e.docStorages, s)

		//初始化Keys存储
		ks, kerr := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.InvertedIndexName, shard)), e.Timeout)
		if kerr != nil {
			panic(err)
		}
		e.invertedIndexStorages = append(e.invertedIndexStorages, ks)

		//id和keys映射
		iks, ikerr := storage.NewStorage(e.getFilePath(fmt.Sprintf("%s_%d", e.Option.PositiveIndexName, shard)), e.Timeout)
		if ikerr != nil {
			panic(ikerr)
		}
		e.positiveIndexStorages = append(e.positiveIndexStorages, iks)
	}
	//go e.automaticGC()

	log.Println("初始化完成")
}

// 自动保存索引，10秒钟检测一次
func (e *Engine) automaticGC() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		<-ticker.C
		//定时GC
		runtime.GC()
	}
}

func (e *Engine) IndexDocument(doc *model.IndexDoc) {
	//根据ID来判断，使用多线程，提速
	e.addDocumentWorkerChan[e.getShard(doc.Id)] <- doc
}

// GetQueue 获取队列剩余
func (e *Engine) GetQueue() int {
	total := 0
	for _, v := range e.addDocumentWorkerChan {
		total += len(v)
	}
	return total
}

// DocumentWorkerExec 添加文档队列
func (e *Engine) DocumentWorkerExec(worker chan *model.IndexDoc) {
	for {
		doc := <-worker
		e.AddDocument(doc)
	}
}

// getShard 计算索引分布在哪个文件块
func (e *Engine) getShard(id uint32) int {
	return int(id % uint32(e.Shard))
}

func (e *Engine) getShardByWord(word string) int {

	return int(utils.StringToInt(word) % uint32(e.Shard))
}

func (e *Engine) InitOption(option *Option) {

	if option == nil {
		//默认值
		option = e.GetOptions()
	}
	e.Option = option
	//shard默认值
	if e.Shard <= 0 {
		e.Shard = 10
	}
	//初始化其他的
	e.Init()

}

func (e *Engine) getFilePath(fileName string) string {
	return e.IndexPath + string(os.PathSeparator) + fileName
}

func (e *Engine) GetOptions() *Option {
	return &Option{
		DocIndexName:      "docs",
		InvertedIndexName: "inverted_index",
		PositiveIndexName: "positive_index",
	}
}

// AddDocument 分词索引
func (e *Engine) AddDocument(index *model.IndexDoc) {
	//等待初始化完成
	e.Wait()
	text := index.Text

	splitWords := e.Tokenizer.Cut(text)

	//id对应的词

	//判断ID是否存在，如果存在，需要计算两次的差值，然后更新
	id := index.Id
	isUpdate := e.optimizeIndex(id, splitWords)

	//没有更新
	if !isUpdate {
		return
	}

	for _, word := range splitWords {
		e.addInvertedIndex(word, id)
	}

	//添加id索引
	e.addPositiveIndex(index, splitWords)
}

// 添加倒排索引
func (e *Engine) addInvertedIndex(word string, id uint32) {
	e.Lock()
	defer e.Unlock()

	shard := e.getShardByWord(word)

	s := e.invertedIndexStorages[shard]

	//string作为key
	key := []byte(word)

	//存在
	//添加到列表
	buf, find := s.Get(key)
	ids := make([]uint32, 0)
	if find {
		utils.Decoder(buf, &ids)
	}

	if !arrays.BinarySearch(ids, id) {
		ids = append(ids, id)
	}

	s.Set(key, utils.Encoder(ids))
}

//	移除没有的词
func (e *Engine) optimizeIndex(id uint32, newWords []string) bool {
	//判断id是否存在
	e.Lock()
	defer e.Unlock()

	//计算差值
	removes, found := e.getDifference(id, newWords)
	if found && len(removes) > 0 {
		//从这些词中移除当前ID
		for _, word := range removes {
			e.removeIdInWordIndex(id, word)
		}
	}

	// 有没有更新
	return !found || len(removes) > 0

}

func (e *Engine) removeIdInWordIndex(id uint32, word string) {

	shard := e.getShardByWord(word)

	wordStorage := e.invertedIndexStorages[shard]

	//string作为key
	key := []byte(word)

	buf, found := wordStorage.Get(key)
	if found {
		ids := make([]uint32, 0)
		utils.Decoder(buf, &ids)

		//移除
		index := arrays.Find(ids, id)
		if index != -1 {
			ids = utils.DeleteArray(ids, index)
			if len(ids) == 0 {
				err := wordStorage.Delete(key)
				if err != nil {
					panic(err)
				}
			} else {
				wordStorage.Set(key, utils.Encoder(ids))
			}
		}
	}

}

// 计算差值
func (e *Engine) getDifference(id uint32, newWords []string) ([]string, bool) {

	shard := e.getShard(id)
	wordStorage := e.positiveIndexStorages[shard]
	key := utils.Uint32ToBytes(id)
	buf, found := wordStorage.Get(key)
	if found {
		oldWords := make([]string, 0)
		utils.Decoder(buf, &oldWords)

		//计算需要移除的
		removes := make([]string, 0)
		for _, word := range oldWords {

			//旧的在新的里面不存在，就是需要移除的
			if !arrays.ArrayStringExists(newWords, word) {
				removes = append(removes, word)
			}
		}
		return removes, true
	}

	return nil, false
}

// 添加正排索引 id=>keys id=>doc
func (e *Engine) addPositiveIndex(index *model.IndexDoc, keys []string) {
	e.Lock()
	defer e.Unlock()

	key := utils.Uint32ToBytes(index.Id)
	shard := e.getShard(index.Id)
	docStorage := e.docStorages[shard]

	//id和key的映射
	positiveIndexStorage := e.positiveIndexStorages[shard]

	doc := &model.StorageIndexDoc{
		IndexDoc: index,
		Keys:     keys,
	}

	//存储id和key以及文档的映射
	docStorage.Set(key, utils.Encoder(doc))

	//设置到id和key的映射中
	positiveIndexStorage.Set(key, utils.Encoder(keys))
}

//求并集
func union(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		if m[v] == 0 {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}

// 求交集
func intersect(slice1, slice2 []string) []string {
	m := make(map[string]int)
	n := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		if m[v] == 1 {
			n = append(n, v)
		}
	}
	return n
}

// 求差集, 此处只统计 s1 对 s2 的差集
func difference(slice1, slice2 []string) []string {
	if len(slice2) == 0 {
		return slice1
	}

	m := make(map[string]int)
	n := make([]string, 0)
	inter := intersect(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}

	for _, value := range slice1 {
		if m[value] == 0 {
			n = append(n, value)
		}
	}
	return n
}

// checkfilter 检查 SliceItems.Id 映射的网页结果是否包含过滤词, true 表示不包含
func (e *Engine) checkfilter(item model.SliceItem, filter []string) bool {
	buf := e.GetDocById(item.Id)
	if buf != nil {
		//gob解析
		storageDoc := new(model.StorageIndexDoc)
		utils.Decoder(buf, &storageDoc)
		//检查是否包含过滤词, 即 storageDoc.Keys 和 filter 交集是否为空
		if len(intersect(storageDoc.Keys, filter)) == 0 {
			return true
		}
	}
	return false
}

// MultiSearch 多线程搜索
func (e *Engine) MultiSearch(request *model.SearchRequest) *model.SearchResult {
	//等待搜索初始化完成
	e.Wait()
	//分词搜索
	query := e.Tokenizer.Cut(request.Query)
	log.Println("分词结果：", query)

	// 对传入的 Filterwords 逐个分词后求并集作为最终的过滤词
	filterwords := make([]string, 0)
	for _, fwords := range request.Filterwords {
		fwslice := e.Tokenizer.Cut(fwords)
		filterwords = union(filterwords, fwslice)
	}
	log.Println("过滤词结果：", filterwords)

	// 将 words 的分词结果经过 filterwords 的过滤，得到最终的搜索词
	// 考虑将 words 的结果对 filterwords 求差集
	words := difference(query, filterwords)
	log.Println("搜索词结果：", words)

	//记录在字典树中
	for _, word := range words {
		e.Recorder.Add(word)
	}

	totalTime := float64(0)

	fastSort := &sorts.FastSort{
		IsDebug: e.IsDebug,
		Order:   request.Order,
	}

	_time := utils.ExecTime(func() {

		base := len(words)
		wg := &sync.WaitGroup{}
		wg.Add(base)

		for _, word := range words {
			go e.processKeySearch(word, fastSort, wg, base)
			//e.processKeySearch(word, fastSort, wg, base)
		}
		wg.Wait()
	})
	if e.IsDebug {
		log.Println("数组查找耗时：", totalTime, "ms")
		log.Println("搜索时间:", _time, "ms")
	}
	// 处理分页
	request = request.GetAndSetDefault()

	//计算交集得分和去重
	fastSort.Process()

	//将去重排序后的搜索结果通过 filterwords 做二次过滤
	SliceItems := fastSort.GetData()
	FilterItems := make([]model.SliceItem, 0)
	for _, sliceitem := range SliceItems {
		if e.checkfilter(sliceitem, filterwords) {
			FilterItems = append(FilterItems, sliceitem)
		}
	}
	fastSort.SetFilters(FilterItems)

	wordMap := make(map[string]bool)
	for _, word := range words {
		wordMap[word] = true
	}

	//读取文档
	var result = &model.SearchResult{
		// Total: fastSort.Count(),
		Total: len(FilterItems),
		Page:  request.Page,
		Limit: request.Limit,
		Words: words,
	}

	_time += utils.ExecTime(func() {

		pager := new(pagination.Pagination)

		// pager.Init(request.Limit, fastSort.Count())
		pager.Init(request.Limit, len(FilterItems))
		//设置总页数
		result.PageCount = pager.PageCount

		//读取单页的id
		if pager.PageCount != 0 {

			start, end := pager.GetPage(request.Page)

			var resultItems = make([]model.SliceItem, 0)
			fastSort.GetAll(&resultItems, start, end)

			count := len(resultItems)

			result.Documents = make([]model.ResponseDoc, count)
			//只读取前面100个
			wg := new(sync.WaitGroup)
			wg.Add(count)
			for index, item := range resultItems {
				go e.getDocument(item, &result.Documents[index], request, &wordMap, wg)
				//e.getDocument(item, &result.Documents[index], request, &wordMap, wg)
			}
			wg.Wait()
		}
	})
	if e.IsDebug {
		log.Println("处理数据耗时：", _time, "ms")
	}
	result.Time = _time

	return result
}

func (e *Engine) getDocument(item model.SliceItem, doc *model.ResponseDoc, request *model.SearchRequest, wordMap *map[string]bool, wg *sync.WaitGroup) {
	buf := e.GetDocById(item.Id)
	defer wg.Done()
	doc.Score = item.Score
	if buf != nil {
		//gob解析
		storageDoc := new(model.StorageIndexDoc)
		utils.Decoder(buf, &storageDoc)
		doc.Document = storageDoc.Document
		doc.Keys = storageDoc.Keys
		text := storageDoc.Text
		//处理关键词高亮
		highlight := request.Highlight
		if highlight != nil {
			//全部小写
			text = strings.ToLower(text)
			//还可以优化，只替换击中的词
			for _, key := range storageDoc.Keys {
				if ok := (*wordMap)[key]; ok {
					text = strings.ReplaceAll(text, key, fmt.Sprintf("%s%s%s", highlight.PreTag, key, highlight.PostTag))
				}
			}
			//放置原始文本
			doc.OriginalText = storageDoc.Text
		}
		doc.Text = text
		doc.Id = item.Id

	}

}

func (e *Engine) processKeySearch(word string, fastSort *sorts.FastSort, wg *sync.WaitGroup, base int) {
	defer wg.Done()

	shard := e.getShardByWord(word)
	log.Println("shard:", shard)
	//读取id
	invertedIndexStorage := e.invertedIndexStorages[shard]
	key := []byte(word)

	buf, find := invertedIndexStorage.Get(key)
	if find {
		ids := make([]uint32, 0)
		//解码
		utils.Decoder(buf, &ids)
		fastSort.Add(&ids)
	}

}

// GetIndexCount 获取索引数量
func (e *Engine) GetIndexCount() int64 {
	var size int64
	for i := 0; i < e.Shard; i++ {
		size += e.invertedIndexStorages[i].GetCount()
	}
	return size
}

// GetDocumentCount 获取文档数量
func (e *Engine) GetDocumentCount() int64 {
	var count int64
	for i := 0; i < e.Shard; i++ {
		count += e.docStorages[i].GetCount()
	}
	return count
}

// GetDocById 通过id获取文档
func (e *Engine) GetDocById(id uint32) []byte {
	shard := e.getShard(id)
	key := utils.Uint32ToBytes(id)
	buf, found := e.docStorages[shard].Get(key)
	if found {
		return buf
	}

	return nil
}

// RemoveIndex 根据ID移除索引
func (e *Engine) RemoveIndex(id uint32) error {
	//移除
	e.Lock()
	defer e.Unlock()

	shard := e.getShard(id)
	key := utils.Uint32ToBytes(id)

	//关键字和Id映射
	//invertedIndexStorages []*storage.LeveldbStorage
	//ID和key映射，用于计算相关度，一个id 对应多个key
	ik := e.positiveIndexStorages[shard]
	keysValue, found := ik.Get(key)
	if !found {
		return errors.New(fmt.Sprintf("没有找到id=%d", id))
	}

	keys := make([]string, 0)
	utils.Decoder(keysValue, &keys)

	//符合条件的key，要移除id
	for _, word := range keys {
		e.removeIdInWordIndex(id, word)
	}

	//删除id映射
	err := ik.Delete(key)
	if err != nil {
		return errors.New(err.Error())
	}

	//文档仓
	err = e.docStorages[shard].Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) Close() {
	e.Lock()
	defer e.Unlock()

	for i := 0; i < e.Shard; i++ {
		e.invertedIndexStorages[i].Close()
		e.positiveIndexStorages[i].Close()
	}
}

// Drop 删除
func (e *Engine) Drop() error {
	e.Lock()
	defer e.Unlock()
	//删除文件
	dir, err := ioutil.ReadDir(e.IndexPath)
	if err != nil {
		return err
	}
	for _, d := range dir {
		err := os.RemoveAll(path.Join([]string{d.Name()}...))
		if err != nil {
			return err
		}
		os.Remove(e.IndexPath)
	}

	//清空内存
	for i := 0; i < e.Shard; i++ {
		e.docStorages = make([]*storage.LeveldbStorage, 0)
		e.invertedIndexStorages = make([]*storage.LeveldbStorage, 0)
		e.positiveIndexStorages = make([]*storage.LeveldbStorage, 0)
	}

	return nil
}
