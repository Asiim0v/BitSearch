package core

import (
	"BitSearch/global"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"gopkg.in/yaml.v2"
)

// Parser 解析器
func Parser() *global.Config {

	var addr = flag.String("addr", "127.0.0.1:5678", "设置监听地址和端口")
	//兼容windows
	dir := fmt.Sprintf(".%sdatabase", string(os.PathSeparator))

	var dataDir = flag.String("data", dir, "设置数据存储目录")

	var debug = flag.Bool("debug", true, "设置是否开启调试模式")

	var dictionaryPath = flag.String("dictionary", "./data/dictionary.txt", "设置词典路径")

	var enableAdmin = flag.Bool("enableAdmin", false, "设置是否开启后台管理")

	var gomaxprocs = flag.Int("gomaxprocs", runtime.NumCPU()*2, "设置GOMAXPROCS")

	var auth = flag.String("auth", "", "开启认证，例如: admin:123456")

	var enableGzip = flag.Bool("enableGzip", true, "是否开启gzip压缩")
	var timeout = flag.Int64("timeout", 10*60, "数据库超时关闭时间(秒)")

	var configPath = flag.String("config", "", "配置文件路径，配置此项其他参数忽略")

	var reminderNum = flag.Int("reminderNum", 5, "默认的输入提示数量")

	var trendNum = flag.Int("trendNum", 10, "默认的热词数量")

	flag.Parse()

	config := &global.Config{}

	if *configPath != "" {
		//解析配置文件
		file, err := ioutil.ReadFile(*configPath)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(file, config)
		if err != nil {
			panic(err)
		}
		return config
	}
	config = &global.Config{
		Addr:        *addr,
		Data:        *dataDir,
		Debug:       *debug,
		Dictionary:  *dictionaryPath,
		EnableAdmin: *enableAdmin,
		Gomaxprocs:  *gomaxprocs,
		Auth:        *auth,
		EnableGzip:  *enableGzip,
		Timeout:     *timeout,
		ReminderNum: *reminderNum,
		TrendNum:    *trendNum,
	}

	return config
}
