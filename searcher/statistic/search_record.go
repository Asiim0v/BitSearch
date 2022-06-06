package statistic

import (
	"BitSearch/searcher/model"
	"container/heap"
	"sort"
)

// 大根堆
type NodeSort []*Node

func (h NodeSort) Len() int {
	return len(h)
}

func (h NodeSort) Less(i, j int) bool {
	return h[i].count > h[j].count
}

func (h *NodeSort) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *NodeSort) Push(x interface{}) {
	*h = append(*h, x.(*Node))
}

func (h *NodeSort) Pop() interface{} {
	res := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return res
}

//Trie根节点
type Trie struct {
	root *Node //根节点,root字段中的node包含所有属于该根结点数据
}

//Node节点
type Node struct {
	char   rune           //字符
	childs map[rune]*Node //所有的子节点用map来存
	Data   interface{}    //自定义数据
	deep   int            //深度
	isTerm bool           //是否是一个字符串的结尾(完整的字符串)
	count  int            //记录一该字符串结尾的个数
	Count  interface{}    //定义数量的接口
	fail   *Node          //Ac自动机的fail节点
}

//构造一个Trie树
func NewTrie() *Trie {
	return &Trie{
		root: NewNode(' ', 1),
	}
}

//创造一个节点,传入字符,深度
func NewNode(char rune, deep int) *Node {
	return &Node{
		char:   char,                     //当前节点的字符
		childs: make(map[rune]*Node, 16), //保存子节点的map
		deep:   deep,                     //深度
		count:  0,                        //数量初始化为0
		fail:   nil,
	}
}

// 关键词加入trie树
func (t *Trie) Add(key string) {

	var parent *Node = t.root
	allChars := []rune(key)

	for _, char := range allChars {
		node, ok := parent.childs[char]
		if !ok {
			node = NewNode(char, parent.deep+1)
			parent.childs[char] = node
		}
		if node.isTerm == true {
			node.count++
			node.Count = node.count
		}
		parent = node
	}
	parent.Data = key
	parent.isTerm = true
	parent.count++
	parent.Count = parent.count
}

//从Trie中查找,前缀搜索
func (t *Trie) prefixSearch(key string, limit int) (nodes NodeSort) {
	// key : 需要搜索的关键词
	// limit : 关键词提示多少个，按照关键词数量显示
	// return : [limit]Node
	heap.Init(&nodes)
	var (
		node  = t.root
		queue []*Node //队列
	)
	// 判断是否在trie树中
	allChars := []rune(key)
	for _, char := range allChars {
		child, ok := node.childs[char]
		if !ok {
			return
		}
		node = child
	}
	queue = append(queue, node)
	for len(queue) > 0 {
		var q2 []*Node
		for _, n := range queue {
			if n.isTerm == true {
				heap.Push(&nodes, n)
			}
			for _, v := range n.childs {
				q2 = append(q2, v)
			}
		}
		queue = q2
	}
	if len(nodes) > limit {
		return nodes[:limit]
	}
	return
}

func (t *Trie) PrefixSearch(key string, limit int) (data []string, sum []int) {
	// return : 关键词， 数量
	nodes := t.prefixSearch(key, limit)
	for _, v := range nodes {
		data = append(data, (*v).Data.(string))
		sum = append(sum, (*v).Count.(int))
	}
	return
}

// GetSearchTrending 获取搜索热度最高的几个关键词
func (t *Trie) GetSearchTrending(limit int) []model.TrendResult {
	nodes := t.prefixSearch("", 100)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Count.(int) > nodes[j].Count.(int) //按照每行的第一个元素排序
	})

	if limit > len(nodes) {
		limit = len(nodes)
	}
	hotSearch := make([]model.TrendResult, limit)
	for i := 0; i < limit; i++ {
		hotSearch[i].Word = nodes[i].Data.(string)
		hotSearch[i].Heat = nodes[i].Count.(int)
	}
	return hotSearch
}
