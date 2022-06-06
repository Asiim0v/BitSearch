package words

import (
	"BitSearch/searcher/utils"
	"bufio"
	"embed"
	"io"
	"log"
	"os"
	"strings"

	"github.com/wangbin/jiebago"
)

var (
	DictionaryFS embed.FS
)

type Tokenizer struct {
	seg jiebago.Segmenter
}

func NewTokenizer(dictionaryPath string) *Tokenizer {
	file, err := DictionaryFS.Open("data/dict.txt")
	if err != nil {
		panic(err)
	}
	utils.ReleaseAssets(file, dictionaryPath)

	tokenizer := &Tokenizer{}

	err = tokenizer.seg.LoadDictionary(dictionaryPath)
	if err != nil {
		panic(err)
	}

	return tokenizer
}

func (t *Tokenizer) Cut(text string) []string {
	//不区分大小写
	text = strings.ToLower(text)
	//移除所有的标点符号
	text = utils.RemovePunctuation(text)
	//移除所有的空格
	text = utils.RemoveSpace(text)

	// wordMap 保存 jiebago 根据词典在搜索引擎模式下的分词结果
	var wordMap = make(map[string]int)

	resultChan := t.seg.CutForSearch(text, true)
	for {
		w, ok := <-resultChan
		if !ok {
			break
		}
		_, found := wordMap[w]
		if !found {
			//去除重复的词
			wordMap[w] = 1
		}
	}

	// 在 wordMap 的基础上加入 stopwords 的过滤, 将过滤结果作为 wordsSlice 并返回
	var stopwords = make(map[string]int)

	FileHandle, err := os.Open("data/stopwords.txt")
	if err != nil {
		log.Println(err)
	}
	defer FileHandle.Close()

	lineReader := bufio.NewReader(FileHandle)
	for {
		line, _, err := lineReader.ReadLine()
		if err == io.EOF {
			break
		}

		_, found := stopwords[string(line)]
		if !found {
			stopwords[string(line)] = 1
		}
	}

	var wordsSlice []string
	for k := range wordMap {
		_, found := stopwords[k]
		if !found {
			wordsSlice = append(wordsSlice, k)
		}
	}

	return wordsSlice
}
