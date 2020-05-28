package common

import (
	"io/ioutil"
	"log"
	"net/http"
)

func Index(arr []string, s string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}

func HTTPGet(u string) ([]byte, error) {
	log.Println("Start request URL:", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
