package utils

import (
	"bufio"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	time2 "time"
)

var IPProxyPool []string //全局Ip代理池子
var IsProxyFlag bool     //是否开启了IP代理

// IpFilesReader IP代理池文件读取
func IpFilesReader(path string) ([]string, error) {
	fileHandle, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	defer fileHandle.Close()

	reader := bufio.NewReader(fileHandle)

	var results []string
	// 按行处理txt
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		results = append(results, string(line))
	}
	return results, nil
}

// CheckProxyIp 代理IP检测
func CheckProxyIp(proxyUrl string) (bool /*是否通过测试*/, string /*响应状态State*/, error) {
	// 代理服务器地址
	url, err := url.Parse(proxyUrl)
	if err != nil {
		return false, "", err
	}

	// 创建带有代理设置的 HTTP 客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(url),
		},
		Timeout: 10 * time2.Second, // 设置超时时间
	}

	// 发起 HTTP 请求
	reqURL := "https://www.baidu.com" // 测试访问，返回请求者 IP
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false, "", err
	}

	// 设置 User-Agent 等头信息（根据需求）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	// 打印响应状态和内容
	//fmt.Println("响应状态:", resp.Status)
	return true, resp.Status, nil
}

// 随机获取代理
func RandProxyStr() string {
	return IPProxyPool[rand.Intn(len(IPProxyPool))]
}
