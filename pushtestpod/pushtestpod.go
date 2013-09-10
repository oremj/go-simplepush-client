package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func SendPing(endPoint string, version int) (err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("PUT", endPoint, strings.NewReader(fmt.Sprintf("version=%d", version)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return
	}
	_, err = client.Do(req)
	return
}

func main() {
	hosts := []string{"est", "test", "fuu"}
	log.Println(hosts)
}
