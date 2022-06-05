package service

import (
	"BitSearch/global"
	"BitSearch/searcher"
	"BitSearch/searcher/model"
	"BitSearch/searcher/system"
	"log"
	"os"
	"runtime"
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

func (b *Base) SearchTrends(database string) []string {
	limit := global.CONFIG.TrendNum
	return b.Container.GetDataBase(database).Recorder.GetSearchTrending(limit)
}
