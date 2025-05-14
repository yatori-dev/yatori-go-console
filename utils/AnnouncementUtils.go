package utils

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// 用于拉取公告
func PullAnnouncement() string {
	url := "https://yatori-dev.github.io/yatori-docs/notice/yatori-go-console-inform.txt"
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// 用于显示公告
func ShowAnnouncement() {
	fmt.Println(PullAnnouncement())
}
