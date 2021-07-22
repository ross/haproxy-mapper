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

func (h *HttpJson) fetch(url string, method string, out interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("user-agent", "haproxy-mapper")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &out)
}
