package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type BasicAuth struct {
	username string
	password string
}

type FilerHandler struct {
	auth     *BasicAuth
	path     string
	timer    *time.Timer
	buffsize int64
	duration int64
}

func NewFilerHandler(username, password, path string, buffsize, duration int64, timer *time.Timer) *FilerHandler {
	return &FilerHandler{
		auth:     &BasicAuth{username: username, password: password},
		path:     path,
		timer:    timer,
		buffsize: buffsize,
		duration: duration,
	}
}

func (filer *FilerHandler) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	if err := r.ParseMultipartForm(filer.buffsize); err != nil {
		http.Error(w, fmt.Sprintf("Parse multipart form error: %v", err), http.StatusBadRequest)
		return
	}
	for _, tmpFiles := range r.MultipartForm.File {
		for _, tmpFile := range tmpFiles {
			if err := os.MkdirAll(name, os.ModePerm); err != nil {
				http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusBadRequest)
				return
			}
			dstFile, err := os.Create(filepath.Join(name, tmpFile.Filename))
			if err != nil {
				http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusBadRequest)
				return
			}
			srcFile, err := tmpFile.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusBadRequest)
				return
			}
			io.Copy(dstFile, srcFile)
		}
	}
}

func (filer *FilerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filer.timer.Reset(time.Second * time.Duration(filer.duration))
	reqUser, reqPass, ok := r.BasicAuth()
	if ok && reqUser == filer.auth.username && reqPass == filer.auth.password {
		name := filepath.Join(filer.path, r.URL.Path[1:])
		switch r.Method {
		case http.MethodGet:
			http.ServeFile(w, r, name)
		case http.MethodPost:
			filer.ServeFile(w, r, name)
		case http.MethodOptions:
			w.Header().Set("Server", "drycc-filer")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
