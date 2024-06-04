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

Filer is mainly a file server, with the main function of uploading and downloading files.

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
