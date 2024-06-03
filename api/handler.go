package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FilerHandler struct {
	path     string
	timer    *time.Timer
	username string
	password string
	buffsize int64
	waittime int64
}

func NewFilerHandler(username, password, path string, buffsize, waittime int64, timer *time.Timer) *FilerHandler {
	return &FilerHandler{
		path:     path,
		timer:    timer,
		username: username,
		password: password,
		buffsize: buffsize,
		waittime: waittime,
	}
}

func (filer *FilerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filer.timer.Reset(time.Second * time.Duration(filer.waittime))
	reqUser, reqPass, ok := r.BasicAuth()
	if ok && reqUser == filer.username && reqPass == filer.password {
		name := filepath.Join(filer.path, r.URL.Path[1:])
		switch r.Method {
		case http.MethodGet:
			http.ServeFile(w, r, name)
		case http.MethodPost:
			if err := r.ParseMultipartForm(filer.buffsize); err == nil {
				for _, tmpFiles := range r.MultipartForm.File {
					for _, tmpFile := range tmpFiles {
						if err := os.MkdirAll(name, os.ModePerm); err == nil {
							dstFile, _ := os.Create(filepath.Join(name, tmpFile.Filename))
							if f, err := tmpFile.Open(); err == nil {
								io.Copy(dstFile, f)
							} else {
								http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusBadRequest)
							}
						}
					}
				}
			} else {
				http.Error(w, fmt.Sprintf("Parse multipart form error: %v", err), http.StatusBadRequest)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
