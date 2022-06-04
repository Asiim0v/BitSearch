package main

import (
	"BitSearch/bootstrap"
	"embed"
)

//go:embed data/*.txt
var dictionaryFS embed.FS

func main() {
	bootstrap.SetupTemplate(dictionaryFS)

	// //初始化容器和参数解析
	// core.Initialize()

	bootstrap.ReadIndex()
}
