package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func getStatusText(code int) string {
	return fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func getIoReadCloserByByteArray(body []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(body))
}

func thisBytesContains(b []byte, s string) bool {
	return strings.Contains(strings.ToLower(string(b)), s)
}

type FastHTTPRequester struct {
	FullUrl string
	Host    string
	Timeout time.Duration
	Url     string
}

func (r *FastHTTPRequester) URL() string {
	if len(r.FullUrl) > 0 {
		return r.FullUrl
	}

	return fmt.Sprintf("%s/%s", strings.TrimRight(r.Host, "/"), strings.TrimLeft(r.Url, "/"))
}

func (r *FastHTTPRequester) NetHttpRequest2FastHttpRequest(req *http.Request) *fasthttp.Request {
	fastRequest := fasthttp.AcquireRequest()
	fastRequest.SetHost(r.Host)
	fastRequest.SetRequestURI(r.URL())

	for key, value := range req.Header {
		switch key {
		case "Content-Type":
			fastRequest.Header.SetContentType(value[0])

			switch value[0] {
			case "application/json", "application/x-www-form-urlencoded", "text/html", "text/plain":
				buf := new(bytes.Buffer)
				buf.ReadFrom(req.Body)
				fastRequest.SetBody(buf.Bytes())

			case "multipart/form-data":
				// TODO: Implement this

			default:
			}
		default:
			fastRequest.Header.Set(key, value[0])
		}
	}
	fastRequest.Header.Set("Accept-Encoding", "gzip")
	fastRequest.Header.SetMethod(req.Method)
	fastRequest.Header.SetUserAgent(req.UserAgent())

	return fastRequest
}

func (r *FastHTTPRequester) FastHttpResponse2NetHttpResponse(fastResponse *fasthttp.Response) *http.Response {
	protocol := []interface{}{"HTTP/1.0", 1, 0}
	if fastResponse.Header.IsHTTP11() {
		protocol = []interface{}{"HTTP/1.1", 1, 1}
	}

	headers := http.Header{}
	fastResponse.Header.VisitAll(func(key, value []byte) {
		if !thisBytesContains(value, "gzip") ||
			!(thisBytesContains(key, "Vary") && thisBytesContains(value, "Accept-Encoding")) {
			headers.Set(string(key), string(value))
		}
	})

	log.Debugf("FastHTTPRequester headers in response: %+v", headers)

	var body io.ReadCloser

	if ce := fastResponse.Header.Peek("Content-Encoding"); thisBytesContains(ce, "gzip") {
		zbody, err := fastResponse.BodyGunzip()
		if err == nil {
			body = getIoReadCloserByByteArray(zbody)
		}
		log.Debugf("FastHTTPRequester response body:\n%v", string(zbody))
	} else if ce := fastResponse.Header.Peek("Content-Encoding"); thisBytesContains(ce, "deflate") {
		ibody, err := fastResponse.BodyInflate()
		if err == nil {
			body = getIoReadCloserByByteArray(ibody)
		}
		log.Debugf("FastHTTPRequester response body:\n%v", string(ibody))
	} else {
		body = getIoReadCloserByByteArray(fastResponse.Body())
		log.Debugf("FastHTTPRequester response body:\n%v", string(fastResponse.Body()))
	}

	response := &http.Response{
		Body:       body,
		Header:     headers,
		Proto:      protocol[0].(string),
		ProtoMajor: protocol[1].(int),
		ProtoMinor: protocol[2].(int),
		Status:     getStatusText(fastResponse.StatusCode()),
		StatusCode: fastResponse.StatusCode(),
	}

	return response
}

func (r *FastHTTPRequester) NetHttp2FastHttpWrapper(req *http.Request) (*http.Response, error) {
	fastRequest := r.NetHttpRequest2FastHttpRequest(req)

	log.Debugf("FastHTTPRequester request\n%s", fastRequest)

	fastResponse := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.DoTimeout(fastRequest, fastResponse, r.Timeout)

	log.Debugf("FastHTTPRequester response\n%s", fastResponse)

	response := r.FastHttpResponse2NetHttpResponse(fastResponse)

	return response, err
}
