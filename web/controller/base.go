package controller

import (
	"BitSearch/searcher/model"

	"github.com/gin-gonic/gin"
)

func Welcome(c *gin.Context) {
	ResponseSuccessWithData(c, "Welcome to BitSearch")
}

// Query 查询
func Query(c *gin.Context) {
	var request = &model.SearchRequest{
		Database: c.Query("database"),
	}
	if err := c.ShouldBind(&request); err != nil {
		ResponseErrorWithMsg(c, err.Error())
		return
	}
	//调用搜索
	r := srv.Base.Query(request)
	ResponseSuccessWithData(c, r)
}

// GC 释放GC
func GC(c *gin.Context) {
	srv.Base.GC()
	ResponseSuccess(c)
}

// Status 获取服务器状态
func Status(c *gin.Context) {
	r := srv.Base.Status()
	ResponseSuccessWithData(c, r)
}

// SearchReminder 搜索提示
func SearchReminder(c *gin.Context) {
	database := c.Query("database")
	query := c.Query("query")
	r := srv.Base.SearchReminder(database, query)
	ResponseSuccessWithData(c, r)
}

// SearchTrends 获取搜索热度
func SearchTrends(c *gin.Context) {
	database := c.Query("database")
	if database == "" {
		ResponseErrorWithMsg(c, "Need Database")
	}
	r := srv.Base.SearchTrends(database)
	ResponseSuccessWithData(c, r)
}

func GetPageDetail(c *gin.Context) {
	url := c.Query("url")
	r := srv.Base.GetDetail(url)
	ResponseSuccessWithData(c, r)
}
