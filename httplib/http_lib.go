package httplib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var Debug bool

type Request struct {
	Method  string
	Path    string
	Query   string
	Headers map[string]string
	BaseUrl string
	Type    string
	Body    *bytes.Reader
	Timeout time.Duration
}

func (req *Request) url() (*url.URL, error) {
	u, err := url.Parse(req.BaseUrl)
	if err != nil {
		return nil, fmt.Errorf("Bad endpoint URL %q: %v", req.BaseUrl, err)
	}
	u.RawQuery = req.Query
	u.Path = req.Path
	return u, nil
}

func initHttpRequest(req *Request) (*http.Request, error) {
	url, _ := req.url()
	newReq, _ := http.NewRequest(req.Method, url.String(), nil)
	for k, v := range req.Headers {
		newReq.Header.Add(k, v)
	}
	if req.Body != nil {
		newReq.Body = ioutil.NopCloser(req.Body)
		newReq.ContentLength = int64(req.Body.Len())
		newReq.Header.Add(CONTENT_LENGTH, fmt.Sprintf("%d", newReq.ContentLength))
		if req.Type != "" {
			newReq.Header.Add(CONTENT_TYPE, req.Type)
		} else {
			newReq.Header.Add(CONTENT_TYPE, OCTET_STREAM)
		}
	}
	return newReq, nil
}

func doHttpRequest(httpClient *http.Client, req *http.Request, res interface{}) (*http.Response, error) {
	result, err := httpClient.Do(req)
	if Debug {
		fmt.Println("+++++++++++++++++++++++++++++++")
		fmt.Println(req)
		fmt.Println("-------------------------------")
		fmt.Println(result)
		fmt.Println("-------------------------------")
		fmt.Println(err)
		fmt.Println("===============================")
	}
	return result, err
}

func Run(req *Request, res interface{}) (*http.Response, error) {
	hreq, _ := initHttpRequest(req)
	httpClient := &http.Client{
		Timeout: req.Timeout,
	}

	return doHttpRequest(httpClient, hreq, res)
}
