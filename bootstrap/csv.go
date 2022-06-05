package bootstrap

import (
	"BitSearch/global"
	"BitSearch/searcher/model"
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
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

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go AddDataset("WebPage", "data/csv/IDCONTENT.csv", wg)
	go AddDataset("Image", "data/csv/WUKONG.csv", wg)
	wg.Wait()
	// AddDatasetWeb("WebPage", "data/csv/IDCONTENT.csv")
	// AddDataset("Image", "data/csv/WUKONG.csv")
}

func AddDataset(name string, filePath string, wg *sync.WaitGroup) {
	defer wg.Done()

	db := global.Container.GetDataBase(name)
	csvfile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer csvfile.Close()

	// Read File
	r := bufio.NewReader(csvfile)
	index := 0

	for {
		utf8_line, err := r.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		// Convert to UTF-8
		if name == "WebPage" {
			utf8_line, _ = ConvertToString(utf8_line, "gbk", "utf-8")
		}
		split_line := strings.Split(utf8_line, ",")
		// sanity check, make sure not go out of bounds
		if len(split_line) < 3 {
			log.Println("leave", len(split_line))
			log.Println(split_line)
			continue
		}

		temp, _ := strconv.ParseUint(split_line[0], 10, 32)
		split_line[2] = strings.TrimRight(split_line[2], "\r\n")

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
		db.IndexDocument(&doc)
	}

	for db.GetQueue() > 0 {
	}
	fmt.Println("index finish")
}

func AddDatasetWeb(name string, filePath string, wg *sync.WaitGroup) {
	defer wg.Done()

	db := global.Container.GetDataBase(name)
	csvfile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer csvfile.Close()

	// Read File
	r := bufio.NewReader(csvfile)
	index := 0

	for {
		utf8_line, err := r.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		// Convert to UTF-8
		utf8_line, _ = ConvertToString(utf8_line, "gbk", "utf-8")
		split_line := strings.Split(utf8_line, ",")
		// sanity check, make sure not go out of bounds
		if len(split_line) < 3 {
			log.Println("leave", len(split_line))
			log.Println(split_line)
			continue
		}

		temp, _ := strconv.ParseUint(split_line[0], 10, 32)
		split_line[2] = strings.TrimRight(split_line[2], "\r\n")

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
		db.IndexDocument(&doc)
	}

	for db.GetQueue() > 0 {
	}
	fmt.Println("index finish")
}
