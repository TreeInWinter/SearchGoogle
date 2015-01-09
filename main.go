package main

import (
	"container/list"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var outputFileName string = "filesName.csv"
var timeout = time.Duration(2 * time.Second)
var ips = list.New()

//超时时间
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

//var c = make(chan string, 100)

/**获得目录的绝对路径****/
func GetFullPath(path string) string {
	absolutePath, _ := filepath.Abs(path)
	fmt.Println("文件的绝对路径是" + absolutePath)
	return absolutePath
}

/***获得该目录下的所有文件名*****/
func PrintFilesName(path string) []string {
	fullPath := GetFullPath(path)

	listStr := list.New()

	filepath.Walk(fullPath, func(path string, fi os.FileInfo, err error) error {
		if nil == fi {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		name := fi.Name()

		if outputFileName != name {
			listStr.PushBack(name)
		}

		return nil
	})
	return ConvertToSlice(listStr)
}

/**文件写入*/
func OutputFilesName(listStr *list.List) {
	files := ConvertToSlice(listStr)
	//sort.StringSlice(files).Sort()// sort

	f, err := os.Create(outputFileName)
	//f, err := os.OpenFile(outputFileName, os.O_APPEND | os.O_CREATE, os.ModeAppend)
	CheckErr(err)
	defer f.Close()

	f.WriteString("\xEF\xBB\xBF")
	writer := csv.NewWriter(f)

	length := len(files)
	for i := 0; i < length; i++ {
		writer.Write([]string{files[i]})
	}

	writer.Flush()
}
func CheckErr(err error) {
	if nil != err {
		panic(err)
	}
}

/*******把list转为数组**********/
func ConvertToSlice(listStr *list.List) []string {
	sli := []string{}
	for el := listStr.Front(); nil != el; el = el.Next() {
		sli = append(sli, el.Value.(string))
	}

	return sli
}

/***
**读取文件
***/
func readFile(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	// fmt.Println(string(fd))
	return string(fd)
}
func testConnection(ip string, channels chan int) {

	transport := http.Transport{
		Dial: dialTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("http://" + ip)
	if err == nil {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println(ip)
			ips.PushBack(ip)
		}

	}
	<-channels
}

func main() {
	//获得文件的绝对位置
	var c = make(chan int, 10)

	var files []string = PrintFilesName("ips")
	var path = GetFullPath("ips")
	for i := 0; i < len(files); i++ {
		if i < 1 {
			var fileFullPath = path + "\\" + files[i]
			contentFile := readFile(fileFullPath)
			var result = strings.Split(contentFile, "\n")
			fmt.Println(len(result))
			for j, value := range result {
				if j < 100 {
					c <- 1
					go testConnection(value, c)

				} else {
					break
				}
			}
		} else {
			break
		}
		//打开文件
	}

	OutputFilesName(ips)
	fmt.Println("done!")
}
