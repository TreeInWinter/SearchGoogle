//批量获取搜狐股票成交明细数据。
package main

import (
	"bufio"
	"container/list"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//var startDate *string = flag.String("d", "Null", "please input a startDate like 20131104")
var num *int = flag.Int("n", 0, "please input a num like 1024")
var ips = list.New()
var outputFileName string = "resutl.csv"

func main() {
	flag.Usage = show_usage
	flag.Parse()
	var (
		stockCodeFile string
		logFileDir    string
		downDir       string
		downFileExt   string
		getUrl        string
	)

	if *num == 0 {
		show_usage()
		return
	}

	cupNum := runtime.NumCPU()
	runtime.GOMAXPROCS(cupNum) //设置cpu的核的数量，从而实现高并发
	c := make(chan int, *num)

	stockCodeFile = "./ips/173.194.0.0_16"

	//日志文件目录，文件下载地址，下载后保存的文件类型

	logFileDir = "./log/" + time.Now().Format("2006-01-02") + "/"
	downDir = "./data/" + time.Now().Format("2006-01-02") + "/"
	downFileExt = ".text"

	if !isDirExists(logFileDir) { //目录不存在，则进行创建。
		err := os.MkdirAll(logFileDir, 777) //递归创建目录，linux下面还要考虑目录的权限设置。
		if err != nil {
			panic(err)
		}
	}
	if !isDirExists(downDir) { //目录不存在，则进行创建。
		err := os.MkdirAll(downDir, 777) //递归创建目录，linux下面还要考虑目录的权限设置。
		if err != nil {
			panic(err)
		}
	}

	logfile, _ := os.OpenFile(logFileDir+time.Now().Format("2006-01-02")+".log", os.O_RDWR|os.O_CREATE, 0)
	logger := log.New(logfile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)

	fh, ferr := os.Open(stockCodeFile)
	if ferr != nil {
		return
	}
	defer fh.Close()
	inputread := bufio.NewReader(fh)

	for i := 1; i <= *num; i++ { //加入goroutine缓冲，4个执行完了再执行下面的4个
		input, _ := inputread.ReadString('\n')
		code := strings.TrimSpace(input)

		getUrl = code
		fmt.Println(getUrl)
		go func(logger *log.Logger, logfile *os.File, code string, downDir string, getUrl string, downFileExt string) {
			testConnection(logger, logfile, code, downDir, getUrl, downFileExt)
			c <- 0
		}(logger, logfile, code, downDir, getUrl, downFileExt)

		if i%4 == 0 { //并发默认为4
			time.Sleep(4 * time.Second) //加入执行缓冲，否则同时发起大量的tcp连接，操作系统会直接返回错误。
		}

	}
	defer logfile.Close()
	for j := 0; j < *num; j++ {
		<-c
	}
	files := ConvertToSlice(ips)
	fmt.Println("符合条数的一共有" + strconv.Itoa(len(files)) + "条")

	GoOutputFiles(ips, downDir)

}

var timeout = time.Duration(2 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
func testConnection(logger *log.Logger, logfile *os.File, code string, downDir string, getUrl string, downFileExt string) {

	transport := http.Transport{
		Dial: dialTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("http://" + code)
	if err == nil {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println(code)
			ips.PushBack(code)
			logger.Println(logfile, code+":http get StatusCode"+strconv.Itoa(resp.StatusCode))
		} else {
			fmt.Println("两秒钟之内没有连接成功")
			logger.Println(logfile, code+":http get StatusCode"+strconv.Itoa(resp.StatusCode))
		}

	} else {

		logger.Println(code + " can do nothing becacuse I can not connect the internet")
	}
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

func show_usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [-n=<num>]\n"+
			"       <command> [<args>]\n\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr,
		"Flags:\n")
	flag.PrintDefaults()

}
func GoOutputFiles(listStr *list.List, downDir string) {
	files := ConvertToSlice(listStr)
	//sort.StringSlice(files).Sort()// sort

	f, err := os.Create(downDir + outputFileName)
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

/*******把list转为数组**********/
func ConvertToSlice(listStr *list.List) []string {
	sli := []string{}
	for el := listStr.Front(); nil != el; el = el.Next() {
		sli = append(sli, el.Value.(string))
	}

	return sli
}

func CheckErr(err error) {
	if nil != err {
		panic(err)
	}
}
