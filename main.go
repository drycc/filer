// Copyright 2024 Drycc Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	pingInterval time.Duration
	bindAddr     string
	lastPingTime time.Time
	pingMutex    sync.RWMutex
)

var rootCmd = &cobra.Command{
	Use:   "pingguard [flags] -- [program args...]",
	Short: "Start any program with ping health check functionality",
	Long: `pingguard is a wrapper program that can:
1. Start any program
2. Provide ping health check functionality, auto-exit if no ping requests received within specified time

Usage examples:
  pingguard --interval=60s --bind=127.0.0.1:8081 -- python -m http.server 8000`,
	Run: runPingguard,
}

func init() {
	rootCmd.Flags().StringVar(&bindAddr, "bind", "127.0.0.1:8081", "ping service bind address and port (format: host:port)")
	rootCmd.Flags().DurationVar(&pingInterval, "interval", 60*time.Second, "ping timeout interval, program will exit if no ping requests received within this time")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runPingguard(_ *cobra.Command, args []string) {
	// Initialize ping time
	updatePingTime()

	// Start ping HTTP server
	server := startPingServer()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("ping server shutdown error: %v", err)
		}
	}()

	// Start target program process
	targetCmd := startTargetProgram(args)
	defer func() {
		if targetCmd.Process != nil {
			log.Println("terminating target process...")
			if err := targetCmd.Process.Signal(os.Interrupt); err != nil {
				log.Printf("failed to send interrupt signal: %v", err)
				_ = targetCmd.Process.Kill()
			}
		}
	}()

	// Start ping checker goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pingChecker(ctx)

	// Wait for signal or process termination
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	procDone := make(chan error, 1)
	go func() {
		procDone <- targetCmd.Wait()
	}()

	select {
	case sig := <-sigChan:
		log.Printf("received signal %v, exiting...", sig)
	case err := <-procDone:
		if err != nil {
			log.Printf("target process exited with error: %v", err)
		} else {
			log.Println("target process exited normally")
		}
	}

	log.Println("program exited")
}

func startPingServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/_/ping", pingHandler)

	server := &http.Server{
		Addr:    bindAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("ping server started on %s", bindAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start ping server: %v", err)
		}
	}()

	return server
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	updatePingTime()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
	log.Printf("received ping request from %s", r.RemoteAddr)
}

func updatePingTime() {
	pingMutex.Lock()
	lastPingTime = time.Now()
	pingMutex.Unlock()
}

func getLastPingTime() time.Time {
	pingMutex.RLock()
	defer pingMutex.RUnlock()
	return lastPingTime
}

func pingChecker(ctx context.Context) {
	ticker := time.NewTicker(pingInterval / 3) // Check frequency is 1/3 of timeout interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Since(getLastPingTime()) > pingInterval {
				log.Printf("no ping requests received for %v, program auto-exiting", pingInterval)
				// Send SIGTERM signal to self
				p, _ := os.FindProcess(os.Getpid())
				_ = p.Signal(syscall.SIGTERM)
				return
			}
		}
	}
}

func startTargetProgram(args []string) *exec.Cmd {
	if len(args) == 0 {
		log.Fatal("error: must provide program and arguments to execute")
	}

	// First argument is program name, rest are arguments
	programName := args[0]
	programArgs := args[1:]

	cmd := exec.Command(programName, programArgs...)

	// Connect standard input/output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	log.Printf("starting command: %s %v", programName, programArgs)
	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	log.Printf("target process started successfully, PID: %d", cmd.Process.Pid)
	return cmd
}
