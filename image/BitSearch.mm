<map version="1.0.1"><node CREATED="1654592550744" ID="ID_root" MODIFIED="1654592550744" TEXT="package main&amp;lt;br&amp;gt;//初始化容器和参数解析&amp;lt;br&amp;gt;core.Initialize()&amp;lt;br&amp;gt;"><node CREATED="1654592550744" ID="ID_2132c65184be" POSITION="right" MODIFIED="1654592550744" TEXT="//&amp;nbsp;注册路由，封装Gin框架&amp;lt;br&amp;gt;&amp;nbsp;r&amp;nbsp;:=&amp;nbsp;router.SetupRouter()"><node CREATED="1654592550744" ID="ID_6ea219b04d0f" MODIFIED="1654592550744" TEXT="Gzip压缩通过中间件形式注册到全局路由&amp;lt;br&amp;gt;"></node><node CREATED="1654592550744" ID="ID_5505a010d405" MODIFIED="1654592550744" TEXT="项目整体路由分组到api路径下"><node CREATED="1654592550744" ID="ID_404371a19e15" MODIFIED="1654592550744" TEXT="基础路由"><node CREATED="1654592550744" ID="ID_5dc1b6fb8649" MODIFIED="1654592550744" TEXT="“api/”—controller.Welcome"><node CREATED="1654592550744" ID="ID_aae1ffb1a291" MODIFIED="1654592550744" TEXT="ResponseSuccessWithData(c,&amp;nbsp;&amp;quot;Welcome&amp;nbsp;to&amp;nbsp;GoFound&amp;quot;)&amp;lt;br&amp;gt;"></node></node><node CREATED="1654592550744" ID="ID_1c6534e534a8" MODIFIED="1654592550744" TEXT="“api/query”—controller.Query"><node CREATED="1654592550744" ID="ID_b2310fbe1bca" MODIFIED="1654592550744" TEXT="底层调用service.Query查询&amp;lt;br&amp;gt;Query首先根据查询请求中的database名称获取或者创建引擎Engine&amp;lt;br&amp;gt;随后调用func&amp;nbsp;(e&amp;nbsp;*Engine)&amp;nbsp;MultiSearch(request&amp;nbsp;*model.SearchRequest)&amp;nbsp;*model.SearchResult多线程查询&amp;lt;br&amp;gt;"><node CREATED="1654592550744" ID="ID_57a7bead43f8" MODIFIED="1654592550744" TEXT="MultiSearch的大致流程如下：&amp;lt;br&amp;gt;1、调用分词器对查询请求中的搜索关键词进行分词&amp;lt;br&amp;gt;2、根据1中的得到的分词数量开启对应数量的goroutine执行&amp;lt;br&amp;gt;processKeySearch并计算全部查询完成所需的时间&amp;lt;br&amp;gt;3、根据查询请求中的参数处理分页，默认查询情况页码为1，每页的结果数为100&amp;lt;br&amp;gt;4、计算得分并去重：&amp;lt;br&amp;gt;本质是将temps中的文档排序后，通过二分查找统计每个文档出现过的次数（对应SliceItem里的Score）&amp;lt;br&amp;gt;最后将去重后的temps放入data中，按照Score的大小排序返回&amp;lt;br&amp;gt;5、计算分页相关的逻辑，得到总分页数，当前页码，当前页的结果数量，搜索关键词等信息放入SearchResult中&amp;lt;br&amp;gt;6、根据6中搜索请求中传入的Page页码获取当前页的结果文档范围，并开启当前页内resultItems对应数量的goroutine，&amp;lt;br&amp;gt;负责把每个SliceItem的信息按照对应格式放入result.Documents中&amp;lt;br&amp;gt;7、统计查询加分页的总处理时间，放入result.Time字段中，并返回result完成整次查询过程"><node CREATED="1654592550744" ID="ID_8e0ef46ad57c" MODIFIED="1654592550744" TEXT="processKeySearch的大致流程如下：&amp;lt;br&amp;gt;1、通过getShardByWord函数获取当前关键词所在的索引文件块&amp;lt;br&amp;gt;（利用了字符串哈希将word映射到整数从而选择对应的索引文件块）&amp;lt;br&amp;gt;2、得到索引文件块的序号shard后，取出对应的倒排索引&amp;lt;br&amp;gt;3、根据倒排索引和关键词word通过LeveldbStorage.Get函数获取倒排索引的value&amp;lt;br&amp;gt;4、如果获取value成功，会将value由[]byte类型解码到[]uint32，并append到FastSort下的temps中&amp;lt;br&amp;gt;（解码操作用到了gob包进行编解码）"></node><node CREATED="1654592550744" ID="ID_93d55d947322" MODIFIED="1654592550744" TEXT="getDocument的大致流程如下：&amp;lt;br&amp;gt;1、GetDocById根据SliceItem的ID从leveldb的文档仓中获取文档（获取文档序号的过程是通过getShard函数计算），&amp;lt;br&amp;gt;注意这里获取的文档是经过gob包编码过的&amp;lt;br&amp;gt;2、对1中获取的结果使用gob包进行解码，并将解码结果放入model.StorageIndexDoc结构中&amp;lt;br&amp;gt;3、给result.Document（底层是model.ResponseDoc，搜索结果中的一个字段）结构体赋值：&amp;lt;br&amp;gt;doc.Score&amp;nbsp;=&amp;nbsp;item.Score&amp;lt;br&amp;gt;doc.Document&amp;nbsp;=&amp;nbsp;storageDoc.Document&amp;lt;br&amp;gt;doc.Keys&amp;nbsp;=&amp;nbsp;storageDoc.Keys&amp;lt;br&amp;gt;doc.OriginalText&amp;nbsp;=&amp;nbsp;storageDoc.Text&amp;lt;br&amp;gt;doc.Text&amp;nbsp;=&amp;nbsp;text（text还额外进行了关键词高亮的处理）&amp;lt;br&amp;gt;doc.Id&amp;nbsp;=&amp;nbsp;item.Id&amp;lt;br&amp;gt;"></node></node></node></node><node CREATED="1654592550745" ID="ID_0c0434c8525b" MODIFIED="1654592550745" TEXT="“api/status”—controller.Status"><node CREATED="1654592550745" ID="ID_834ec71369b2" MODIFIED="1654592550745" TEXT="底层调用service.Status获取服务器状态&amp;lt;br&amp;gt;返回一些memory，cpu，disk，system相关的信息"></node></node><node CREATED="1654592550745" ID="ID_4027d3bf16fe" MODIFIED="1654592550745" TEXT="“api/gc”—controller.GC"><node CREATED="1654592550745" ID="ID_46e674c6e4a8" MODIFIED="1654592550745" TEXT="底层调用service.GC --&amp;gt; runtime.GC() 显示调用垃圾回收&amp;lt;br&amp;gt;"></node></node><node CREATED="1654592550745" ID="ID_f2166111d499" MODIFIED="1654592550745" TEXT="&amp;quot;api/trend&amp;quot; —controller.SearchTrends&amp;lt;br&amp;gt;"><node CREATED="1654592550745" ID="ID_4de62a9de2fe" MODIFIED="1654592550745" TEXT="调用service.SeaerchTrend()&amp;lt;br&amp;gt;"><node CREATED="1654592550745" ID="ID_0fff5a8311fd" MODIFIED="1654592550745" TEXT="从Container Recorder（字典树）中获得词频并统计出高频词汇返回"></node></node></node><node CREATED="1654592550745" ID="ID_9e5b777fb60b" MODIFIED="1654592550745" TEXT="&amp;quot;api/reminder&amp;quot; —controller.SearchReminder&amp;lt;br&amp;gt;"><node CREATED="1654592550745" ID="ID_4c72671625b5" MODIFIED="1654592550745" TEXT="调用service.SearchReminder"><node CREATED="1654592550745" ID="ID_db054e134fdb" MODIFIED="1654592550745" TEXT="对Container Recoder(字典树)进行检索，查找相关的节点并记录，直到访问叶节点"></node></node></node></node><node CREATED="1654592550745" ID="ID_263ff2c70164" MODIFIED="1654592550745" TEXT="索引路由"><node CREATED="1654592550745" ID="ID_090188ec06b9" MODIFIED="1654592550745" TEXT="“api/index”—controller.AddIndex"></node><node CREATED="1654592550745" ID="ID_176fb67342ac" MODIFIED="1654592550745" TEXT="“api/index/batch”—controller.BatchAddIndex"><node CREATED="1654592550745" ID="ID_ffaee8f6ffac" MODIFIED="1654592550745" TEXT="批量添加索引底层调用service.BatchAddIndex，和上面单条索引一样需要传入dbname和IndexDoc切片&amp;lt;br&amp;gt;底层调用的&amp;lt;b&amp;gt;同样是Engine.IndexDocument函数&amp;lt;/b&amp;gt;，只是遍历传入的IndexDoc切片，每个IndexDoc开启一个goroutine插入"></node></node><node CREATED="1654592550745" ID="ID_337f6be10402" MODIFIED="1654592550745" TEXT="“api/index/remove”—controller.RemoveIndex"><node CREATED="1654592550745" ID="ID_fa76b1d9772a" MODIFIED="1654592550745" TEXT="删除索引底层调用service.RemoveIndex，同样需要传入dbname和removeIndexModel结构体（该结构体仅含有一个索引ID）&amp;lt;br&amp;gt;通过dbname获取对应Engine后，&amp;lt;b&amp;gt;调用Engine.RemoveIndex函数&amp;lt;/b&amp;gt;，根据ID删除索引&amp;lt;br&amp;gt;"><node CREATED="1654592550745" ID="ID_bfa99fd1176f" MODIFIED="1654592550745" TEXT="RemoveIndex大致流程如下：&amp;lt;br&amp;gt;1、通过传入的ID先从对应的正排索引中查到对应的words []byte，对于words中每个word，调用&amp;lt;br&amp;gt;Engine.removeIdInWordIndex函数将word对应的倒排索引中的ID移除&amp;lt;br&amp;gt;2、将正排索引positiveIndexStorages[shard]和文档仓docStorages[shard]对应的ID记录删除&amp;lt;br&amp;gt;"></node></node></node></node><node CREATED="1654592550745" ID="ID_a3ac6c48defc" MODIFIED="1654592550745" TEXT="数据库路由"><node CREATED="1654592550745" ID="ID_9a51db1f9fe4" MODIFIED="1654592550745" TEXT="“api/db/list”—controller.DBS"><node CREATED="1654592550745" ID="ID_5129a4c6276b" MODIFIED="1654592550745" TEXT="srv.Database.Show() --&amp;gt;&amp;nbsp;Container.GetDataBases()打印当前container内已经创建的所有engine，每个engine对应一个database&amp;lt;br&amp;gt;并返回一个包含所有engine的ResponseData格式的Json响应&amp;lt;br&amp;gt;"></node></node><node CREATED="1654592550745" ID="ID_e935f4c5668f" MODIFIED="1654592550745" TEXT="“api/db/drop”—controller.DatabaseDrop"><node CREATED="1654592550745" ID="ID_570b4bc00752" MODIFIED="1654592550745" TEXT="删除数据库和上述索引部分的路由一样，&amp;lt;b&amp;gt;需要在请求URL中加入?database=db_name&amp;lt;/b&amp;gt;，如果没有指定dbname在程序内部会报错&amp;lt;br&amp;gt;srv.Database.Drop(dbName) --&amp;gt;&amp;nbsp;Container.DropDataBase(dbName) --&amp;gt;&amp;nbsp;engines[name].Drop()&amp;lt;br&amp;gt;删除数据存储目录（global.CONFIG.Data）下的对应数据库目录内的所有文件，并清空内存&amp;lt;br&amp;gt;"></node></node><node CREATED="1654592550745" ID="ID_10b4590b3cdb" MODIFIED="1654592550745" TEXT="“api/db/create”—controller.DatabaseCreate"><node CREATED="1654592550745" ID="ID_fcc9c8d9ff35" MODIFIED="1654592550745" TEXT="加入dbname，srv.Database.Create(dbName) --&amp;gt;&amp;nbsp;Container.GetDataBase(dbName)&amp;lt;br&amp;gt;GetDataBase函数在前面已经多次出现过了，在索引路由、基础路由的Query等地方都有应用，本质就是根据dbname获取或创建数据库&amp;lt;br&amp;gt;"></node></node></node><node CREATED="1654592550745" ID="ID_616bfb211069" MODIFIED="1654592550745" TEXT="分词路由"><node CREATED="1654592550745" ID="ID_24787b113506" MODIFIED="1654592550745" TEXT="“api/word/cut”—controller.WordCut"><node CREATED="1654592550745" ID="ID_4e5d9db3172d" MODIFIED="1654592550745" TEXT="在URL中加入&amp;lt;b&amp;gt;?q=&amp;quot;需要分词的内容&amp;quot;&amp;lt;/b&amp;gt;，调用程序初始化的分词器Tokenizer.Cut返回分词结果，分词器根据设置的词典地址进行初始化（默认./data/dictionary.txt）&amp;lt;br&amp;gt;"></node></node></node></node></node><node CREATED="1654592550745" ID="ID_357b9cd82c6b" POSITION="right" MODIFIED="1654592550745" TEXT="开启一个goroutine启动http服务&amp;lt;br&amp;gt;&amp;lt;br&amp;gt;"></node><node CREATED="1654592550745" ID="ID_70d6b2d02671" POSITION="right" MODIFIED="1654592550745" LINK="https://github.com/newpanjing/gofound/blob/main/docs/config.md" TEXT="core.Parser()&amp;lt;br&amp;gt;// 初始化global.Config&amp;lt;br&amp;gt;flag包提供自定义命令行初始化&amp;lt;br&amp;gt;"></node><node CREATED="1654592550745" ID="ID_52fbf8423f96" POSITION="right" MODIFIED="1654592550745" TEXT="//初始化分词器&amp;lt;br&amp;gt;&amp;nbsp;tokenizer&amp;nbsp;:=&amp;nbsp;NewTokenizer(global.CONFIG.Dictionary)&amp;lt;br&amp;gt;&amp;nbsp;global.Container&amp;nbsp;=&amp;nbsp;NewContainer(tokenizer)"></node><node CREATED="1654592550745" ID="ID_9eb8c31ca977" POSITION="right" MODIFIED="1654592550745" TEXT="//&amp;nbsp;初始化业务逻辑&amp;lt;br&amp;gt;&amp;nbsp;controller.NewServices()"><node CREATED="1654592550745" ID="ID_5226b60292a5" MODIFIED="1654592550745" TEXT="Base基本功能&amp;lt;br&amp;gt;"><node CREATED="1654592550745" ID="ID_57ed5a1d6814" MODIFIED="1654592550745" TEXT="Query 查询"></node><node CREATED="1654592550745" ID="ID_4336002229b7" MODIFIED="1654592550745" TEXT="GC 释放GC"></node><node CREATED="1654592550745" ID="ID_5d69ae45b50b" MODIFIED="1654592550745" TEXT="Status 获取服务器状态"></node><node CREATED="1654592550745" ID="ID_eb41d733a5f5" MODIFIED="1654592550745" TEXT="Restart 重启服务"></node></node><node CREATED="1654592550745" ID="ID_dbaedadc19ea" MODIFIED="1654592550745" TEXT="Index索引功能"><node CREATED="1654592550745" ID="ID_f8172a602262" MODIFIED="1654592550745" TEXT="AddIndex 添加索引"></node><node CREATED="1654592550745" ID="ID_8370f870ba46" MODIFIED="1654592550745" TEXT="BatchAddIndex 批次添加索引"></node><node CREATED="1654592550745" ID="ID_9dfdd37bbaed" MODIFIED="1654592550745" TEXT="RemoveIndex 删除索引"></node></node><node CREATED="1654592550745" ID="ID_eabbab4833d1" MODIFIED="1654592550745" TEXT="Database数据库功能"><node CREATED="1654592550745" ID="ID_d5839e2f2465" MODIFIED="1654592550745" TEXT="Show 查看数据库"></node><node CREATED="1654592550745" ID="ID_f7dfa4923947" MODIFIED="1654592550745" TEXT="Drop 删除数据库"></node><node CREATED="1654592550745" ID="ID_5a15b421d1da" MODIFIED="1654592550745" TEXT="Create 创建数据库"></node></node><node CREATED="1654592550745" ID="ID_20cf591a9d79" MODIFIED="1654592550745" TEXT="Word 分词功能"><node CREATED="1654592550745" ID="ID_8c9d3dad143f" MODIFIED="1654592550745" TEXT="WordCut 分词"></node></node></node><node CREATED="1654592550745" ID="ID_c1786d72ac72" POSITION="right" MODIFIED="1654592550745" TEXT="返回的Json响应数据结构&amp;lt;br&amp;gt;"><node CREATED="1654592550745" ID="ID_95b826ad64b9" MODIFIED="1654592550745" TEXT="ResponseSuccessWithData为gofound封装的一个携带数据成功返回的响应函数&amp;lt;br&amp;gt;状态码200，具体数据为：&amp;lt;br&amp;gt;{&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;State:&amp;nbsp;&amp;nbsp;&amp;nbsp;true,&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;Message:&amp;nbsp;&amp;quot;success&amp;quot;,&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;Data:&amp;nbsp;&amp;nbsp;&amp;nbsp;&amp;nbsp;data,&amp;lt;br&amp;gt;&amp;nbsp;}"></node><node CREATED="1654592550745" ID="ID_11068708ee6b" MODIFIED="1654592550745" TEXT="ResponseErrorWithMsg为gofound封装的一个返回错误的响应函数&amp;lt;br&amp;gt;状态码200，具体数据为：&amp;lt;br&amp;gt;{&amp;lt;br&amp;gt;&amp;nbsp; State:&amp;nbsp;&amp;nbsp;&amp;nbsp;false,&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;Message:&amp;nbsp;message,&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;Data:&amp;nbsp;&amp;nbsp;&amp;nbsp;&amp;nbsp;nil,&amp;lt;br&amp;gt;&amp;nbsp;}"></node><node CREATED="1654592550745" ID="ID_49d2305882f8" MODIFIED="1654592550745" TEXT="ResponseSuccess为gofound封装的一个返回成功的响应函数&amp;lt;br&amp;gt;状态码200，具体数据为：&amp;lt;br&amp;gt;{&amp;lt;br&amp;gt;&amp;nbsp; State:&amp;nbsp;&amp;nbsp;&amp;nbsp;true,&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;Message:&amp;nbsp;&amp;quot;success&amp;quot;,&amp;lt;br&amp;gt;&amp;nbsp;&amp;nbsp;Data:&amp;nbsp;&amp;nbsp;&amp;nbsp;&amp;nbsp;nil,&amp;lt;br&amp;gt;&amp;nbsp;}"></node></node><node CREATED="1654592550745" ID="ID_85ac17b13d3a" POSITION="right" MODIFIED="1654592550745" TEXT="搜索时输入接口的Json格式请求"></node><node CREATED="1654592550745" ID="ID_3b4fff422b3b" POSITION="right" MODIFIED="1654592550745" TEXT="搜索返回的Json格式的结果"></node><node CREATED="1654592550745" ID="ID_45a09f15a861" POSITION="right" MODIFIED="1654592550745" TEXT="全局容器"></node><node CREATED="1654592550745" ID="ID_755a1147e0b8" POSITION="right" MODIFIED="1654592550745" TEXT="初始配置"></node><node CREATED="1654592550745" ID="ID_56e9bfb0c5b6" POSITION="right" MODIFIED="1654592550745" TEXT="数据库引擎"></node></node></map>