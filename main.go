package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

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

var timeout = time.Duration(2 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

var c = make(chan int, 100)

func main() {
	defer close(c)

	contentFile := readFile("F:\\go\\test\\ips\\173.194.0.0_16")
	var result = strings.Split(contentFile, "\n")
	fmt.Println(len(result))
	for i, value := range result {
		if i < 1000 {
			c <- 1
			go testConnection(value)
		} else {
			break
		}

	}

	fmt.Println("done!")
}
func testConnection(ip string) {
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
		}

	}
	<-c
}
