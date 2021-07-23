package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpJson struct {
	client http.Client
}

func HttpJsonCreate() HttpJson {
	return HttpJson{
		client: http.Client{
			Timeout: time.Duration(10 * time.Second),
		},
	}
}

func (h *HttpJson) FetchBody(url string, method string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("user-agent", "haproxy-mapper")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (h *HttpJson) Fetch(url string, method string, out interface{}) error {
	body, err := h.FetchBody(url, method)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &out)
}
