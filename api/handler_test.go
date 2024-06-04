package api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func runServer() *httptest.Server {
	timer := time.NewTimer(time.Second * time.Duration(3600))
	server := httptest.NewServer(NewFilerHandler("drycc", "drycc", "/tmp", 32<<16, 3600, timer))
	return server
}

func newRequest(uri string, name, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(name, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("drycc", "drycc")
	return req, err
}

func TestFile(t *testing.T) {
	t.Parallel()
	server := runServer()
	defer server.Close()
	client := server.Client()

	_, path, _, _ := runtime.Caller(0)
	req, err := newRequest(fmt.Sprintf("%s/aaa/bbb", server.URL), "file", path)
	if err != nil {
		log.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		t.Fatal(fmt.Errorf("post file error: %d, %s", res.StatusCode, string(body)))
	}
	defer res.Body.Close()
	// test get
	url := fmt.Sprintf("%s/aaa/bbb/handler_test.go", server.URL)
	req, err = http.NewRequest("GET", url, nil)
	req.SetBasicAuth("drycc", "drycc")
	if err != nil {
		log.Fatal(err)
	}
	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		t.Fatal(fmt.Errorf("post file error: %d, %s", res.StatusCode, string(body)))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	expect, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	if string(body) != string(expect) {
		log.Fatalf("expect: %s\r\nactual: %s", string(body), string(expect))
	}
}
