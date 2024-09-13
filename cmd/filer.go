package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/drycc/filer/api"
)

func main() {
	bind := flag.String("bind", ":8100", "port to serve on")
	path := flag.String("path", ".", "the directory of file to host")
	buffsize := flag.Int64("buffsize", 32<<16, "memory size for multipart form")
	duration := flag.Int64("duration", 3600, "start duration for filer")
	waittime := flag.Int64("waittime", 1200, "close waittime for filer")
	username := flag.String("username", "drycc", "provide a user for the account")
	password := flag.String("password", "drycc", "provide a pass for the account")

	flag.Parse()
	log.Printf("Serving %s on http://%s\n", *path, *bind)
	timer := time.NewTimer(time.Second * time.Duration(*duration))
	server := &http.Server{
		Addr:           *bind,
		Handler:        api.NewFilerHandler(*username, *password, *path, *buffsize, *duration, timer),
		ReadTimeout:    time.Duration(*duration) * time.Second,
		WriteTimeout:   time.Duration(*duration) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	<-timer.C
	log.Println("Shutting down server after requests finish..")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*waittime))
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting now")
}
