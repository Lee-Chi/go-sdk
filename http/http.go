package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func NewRequest() *Request {
	return &Request{
		header: map[string]string{},
	}
}

type Request struct {
	header map[string]string
}

func (r *Request) SetHeader(key string, value string) *Request {
	r.header[key] = value
	return r
}

func (r Request) Get(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 300 * time.Second,
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range r.header {
		request.Header.Set(k, v)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code is %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (r Request) PostJson(url string, params interface{}) ([]byte, error) {
	client := http.Client{
		Timeout: 300 * time.Second,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}

	for k, v := range r.header {
		request.Header.Set(k, v)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("response status code is %d, and read response body failed, %v, body: %v", response.StatusCode, err, response.Body)
		}

		return nil, fmt.Errorf("response status code is %d, and response body is %s", response.StatusCode, string(responseBody))
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("response status code is %d, but read response body failed, %v, body: %v", response.StatusCode, err, response.Body)
	}

	return responseBody, nil
}
