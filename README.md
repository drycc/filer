# Drycc Filer

[![Build Status](https://woodpecker.drycc.cc/api/badges/drycc/filer/status.svg)](https://woodpecker.drycc.cc/drycc/filer)
[![codecov](https://codecov.io/gh/drycc/filer/graph/badge.svg?token=mdHMyJabMy)](https://codecov.io/gh/drycc/filer)
[![Go Report Card](https://goreportcard.com/badge/github.com/drycc/filer)](https://goreportcard.com/report/github.com/drycc/filer)
[![codebeat badge](https://codebeat.co/badges/753e5188-2ffa-4f43-b8b4-340166a2b98c)](https://codebeat.co/projects/github-com-drycc-filer-main)

Drycc - A Fork of Drycc Workflow

Drycc (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes][k8s-home] cluster, making it easy to deploy and manage applications on your own servers.

For more information about Drycc Workflow, please visit the main project page at [drycc workflow][workflow].

We welcome your input! If you have feedback, please [submit an issue][issues]. If you'd like to participate in development, please read the "Development" section below and [submit a pull request][prs].

# About

Filer is a specialized wrapper for rclone services with an automatic exit mechanism that provides the following key features:

1. **Start rclone serve**: Launches rclone serve commands with various backends and protocols
2. **Ping health check**: Provides a `/_/ping` endpoint and automatically exits if no ping requests are received within the specified time period

This tool is specifically designed for rclone file serving scenarios in containerized environments, where automatic cleanup is essential when the service is no longer needed.

## Usage

```bash
filer [flags] -- rclone serve [rclone args...]
```

### Flags

- `--interval`: Ping timeout interval, program will exit if no ping requests received within this time (default: 60s)
- `--bind`: Ping service bind address and port in format host:port (default: 127.0.0.1:8081)

### Examples

```bash
# Start rclone HTTP server with ping health check (60s timeout, bind to 127.0.0.1:8081)
filer -- rclone serve http /path/to/files

# Start rclone WebDAV server with custom ping settings
filer --interval=30s --bind=0.0.0.0:8080 -- rclone serve webdav /path/to/files --addr :8000

# Start rclone FTP server with longer timeout
filer --interval=300s --bind=:9000 -- rclone serve ftp /path/to/files --addr :2121

# Start rclone with remote storage (e.g., S3)
filer --interval=120s -- rclone serve http s3:mybucket/folder --addr :8080

# Send ping request to keep service alive
curl http://127.0.0.1:8081/_/ping
```

If no ping requests are received within the specified `--interval` time, both the filer wrapper and the rclone service will automatically shut down. This is particularly useful for temporary file sharing scenarios or containerized applications that need automatic cleanup when no longer in use.

# Development

The Drycc project welcomes contributions from all developers. The high level process for development matches many other open source projects. See below for an outline.

* Fork this repository
* Make your changes
* [Submit a pull request][prs] (PR) to this repository with your changes, and unit tests whenever possible
	* If your PR fixes any [issues][issues], make sure you write `Fixes #1234` in your PR description (where `#1234` is the number of the issue you're closing)
* The Drycc core contributors will review your code. After each of them sign off on your code, they'll label your PR with `LGTM1` and `LGTM2` (respectively). Once that happens, a contributor will merge it

## Container Based Development Environment

The preferred environment for development uses [the `go-dev` Container image](https://github.com/drycc/go-dev). The tools described in this section are used to build, test, package and release each version of Drycc.

To use it yourself, you must have [make](https://www.gnu.org/software/make/) installed and Container installed and running on your local development machine.

If you don't have Podman installed, please go to https://podman.io/ to install it.

After you have those dependencies, build your code with `make build` and execute unit tests with `make test`.


## Dogfooding

Please follow the instructions on the [official Drycc docs](http://www.drycc.cc/docs) to install and configure your Drycc Workflow cluster and all related tools, and deploy and configure an app on Drycc Workflow.

[prs]: https://github.com/drycc/filer/pulls
[issues]: https://github.com/drycc/filer/issues
[workflow]: https://github.com/drycc/workflow
[k8s-home]: https://github.com/kubernetes/kubernetes
