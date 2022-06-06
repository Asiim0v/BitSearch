package service

import (
	"BitSearch/global"
	"BitSearch/searcher"
	"BitSearch/searcher/model"
	"BitSearch/searcher/system"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// Base 基础管理
type Base struct {
	Container *searcher.Container
	Callback  func() map[string]interface{}
}

func NewBase() *Base {
	return &Base{
		Container: global.Container,
		Callback:  Callback,
	}
}

// Query 查询
func (b *Base) Query(request *model.SearchRequest) *model.SearchResult {
	log.Println("query:", request)
	log.Println("query_db:", request.Database)
	log.Println("filterwords:", request.Filterwords)
	return b.Container.GetDataBase(request.Database).MultiSearch(request)
}

// GC 释放GC
func (b *Base) GC() {
	runtime.GC()
}

// Status 获取服务器状态
func (b *Base) Status() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s := b.Callback()

	r := map[string]interface{}{
		"memory": system.GetMemStat(),
		"cpu":    system.GetCPUStatus(),
		"disk":   system.GetDiskStat(),
		"system": s,
	}
	return r
}

// Restart 重启服务
func (b *Base) Restart() {
	// TODD 未实现
	os.Exit(0)
}

func (b *Base) SearchReminder(database string, query string) []string {
	limit := global.CONFIG.ReminderNum
	data, _ := b.Container.GetDataBase(database).Recorder.PrefixSearch(query, limit)
	return data
}

func (b *Base) SearchTrends(database string) []model.TrendResult {
	limit := global.CONFIG.TrendNum
	return b.Container.GetDataBase(database).Recorder.GetSearchTrending(limit)
}

func (b *Base) GetDetail(url string) (ret []string) {
	resp, _ := http.Get(url)

	data, _ := ioutil.ReadAll(resp.Body)
	if !utf8.Valid(data) {
		data, _ = simplifiedchinese.GBK.NewDecoder().Bytes(data)
	}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(data))

	title := doc.Find("title").Text()
	content := doc.Find("div").Text()
	content = strings.Replace(content, "\n", "", -1)
	content = strings.Replace(content, "\t", "", -1)
	content = strings.Replace(content, " ", "", -1)

	ret = append(ret, title)
	ret = append(ret, content[:80])

	return
}
