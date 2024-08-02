package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
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

var (
	ErrNotFound     = fmt.Errorf("not found")
	ErrUnauthorized = fmt.Errorf("unauthorized")
)

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
		if response.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		} else if response.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}

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
		if response.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		} else if response.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}

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

func copyFile(part io.Writer, header *multipart.FileHeader) error {
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	return nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (r Request) PostMultipart(url string, files map[string]*multipart.FileHeader, texts map[string]string) ([]byte, error) {
	client := http.Client{
		Timeout: 300 * time.Second,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range texts {
		writer.WriteField(k, v)
	}

	for k, v := range files {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(k), escapeQuotes(v.Filename)))
		h.Set("Content-Type", v.Header.Get("Content-Type"))
		part, err := writer.CreatePart(v.Header)
		if err != nil {
			return nil, fmt.Errorf("create form field failed, %v", err)
		}

		if err := copyFile(part, v); err != nil {
			return nil, fmt.Errorf("copy file failed, %v", err)
		}
	}

	contentType := writer.FormDataContentType()

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer failed, %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range r.header {
		request.Header.Set(k, v)
	}
	request.Header.Set("Content-Type", contentType)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		} else if response.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}

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
