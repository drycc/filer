package api

import (
	"encoding/json"
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

func (filer *FilerHandler) postFile(w http.ResponseWriter, r *http.Request, name string) {
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
			defer dstFile.Close()

			srcFile, err := tmpFile.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusBadRequest)
				return
			}
			defer srcFile.Close()
			io.Copy(dstFile, srcFile)
		}
	}
}

func (filer *FilerHandler) deleteFile(w http.ResponseWriter, name string) {
	err := os.RemoveAll(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusInternalServerError)
	}
}

func (filer *FilerHandler) listDir(w http.ResponseWriter, name string) {
	result := make([]map[string]string, 0)

	if files, err := os.ReadDir(name); err != nil {
		if f, err := os.Stat(name); err != nil {
			http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
			return
		} else if !f.IsDir() {
			item := make(map[string]string)
			item["name"] = f.Name()
			item["type"] = "file"
			item["size"] = fmt.Sprint(f.Size())
			item["timestamp"] = f.ModTime().Format(time.RFC3339)
			result = append(result, item)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		for _, file := range files {
			item := make(map[string]string)
			if info, err := file.Info(); err == nil {
				item["name"] = info.Name()
				if info.IsDir() {
					item["type"] = "dir"
				} else {
					item["type"] = "file"
				}
				item["size"] = fmt.Sprint(info.Size())
				item["timestamp"] = info.ModTime().Format(time.RFC3339)
			}
			result = append(result, item)
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (filer *FilerHandler) getFile(w http.ResponseWriter, name string) {
	file, err := os.Stat(name)
	if os.IsNotExist(err) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}
	if file.IsDir() {
		http.Error(w, fmt.Sprintf("Path %s a directory, not a file", name), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprint(file.Size()))
	if f, err := os.Open(name); err == nil {
		io.Copy(w, f)
		defer f.Close()
	}
}

func (filer *FilerHandler) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	action := r.URL.Query().Get("action")
	switch action {
	case "get":
		filer.getFile(w, name)
	case "list":
		filer.listDir(w, name)
	default:
		http.Error(w, fmt.Sprintf("Unsupported action %s", action), http.StatusBadRequest)
	}
}

func (filer *FilerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filer.timer.Reset(time.Second * time.Duration(filer.duration))
	reqUser, reqPass, ok := r.BasicAuth()
	if ok && reqUser == filer.auth.username && reqPass == filer.auth.password {
		name := filepath.Join(filer.path, r.URL.Path[1:])
		switch r.Method {
		case http.MethodGet:
			filer.serveFile(w, r, name)
		case http.MethodPost:
			filer.postFile(w, r, name)
		case http.MethodDelete:
			filer.deleteFile(w, name)
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
