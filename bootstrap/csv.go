package bootstrap

import (
	"BitSearch/searcher"
	"BitSearch/searcher/model"
	"BitSearch/searcher/words"
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
)

func ConvertToString(src string, srcCode string, tagCode string) (string, error) {
	if len(src) == 0 || len(srcCode) == 0 || len(tagCode) == 0 {
		return "", errors.New("input arguments error")
	}
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)

	return result, nil
}

// ReadIndex read each line from csv file and convert it to model.Index
func ReadIndex() {
	// compute time
	start := time.Now()
	defer func() {
		cost := time.Since(start)
		fmt.Println("cost=", cost)
	}()

	// test
	tokenizer := words.NewTokenizer("../searcher/words/data/dictionary.txt")

	var engine = &searcher.Engine{
		IndexPath: "./test/index/db2",
		Tokenizer: tokenizer,
	}
	option := engine.GetOptions()

	engine.InitOption(option)

	// Open the file
	csvfile, err := os.Open("data/csv/IDCONTENT.csv")
	if err != nil {
		panic(err)
	}
	defer csvfile.Close()

	// Read File
	r := bufio.NewReader(csvfile)
	index := 0

	for {
		gbk_line, err := r.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		// Convert to UTF-8
		utf8_line, _ := ConvertToString(gbk_line, "gbk", "utf-8")
		split_line := strings.Split(utf8_line, ",")
		// sanity check, make sure not go out of bounds
		if len(split_line) != 3 {
			continue
		}

		temp, _ := strconv.ParseUint(split_line[0], 10, 32)

		if index%1000 == 0 {
			fmt.Println(index)
		}
		index++

		data := make(map[string]interface{})
		data["URL"] = split_line[2]
		data["cid"] = uint32(temp)
		data["title"] = split_line[1]

		doc := model.IndexDoc{
			Id:       uint32(temp),
			Text:     split_line[1],
			Document: data,
		}
		engine.IndexDocument(&doc)
	}

	for engine.GetQueue() > 0 {
	}
	fmt.Println("index finish")
}
