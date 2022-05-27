package logger

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type HttpClient struct {
	client *http.Client
}

var once sync.Once
var hpool *HttpClient

// NewHttpClient
func NewHttpClient(max_conn_per, max_idle_conn_per int, duration int64) *HttpClient {
	once.Do(func() {
		hpool = new(HttpClient)
		hpool.client = &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:     max_conn_per,
				MaxIdleConnsPerHost: max_idle_conn_per,
			},
			Timeout: time.Duration(duration) * time.Second,
		}
	})
	return hpool
}

// send a http request of post or get
func (h *HttpClient) Request(url string, method string, data string, header map[string]string) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}

	for h, v := range header {
		req.Header.Set(h, v)
	}

	response, err := h.client.Do(req)
	if err != nil {
		return "", err
	} else if response != nil {
		defer response.Body.Close()

		r_body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		} else {
			return string(r_body), nil
		}
	} else {
		return "", nil
	}
}
