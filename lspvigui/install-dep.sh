#!/usr/bin/env bash
export GO111MODULE=off; go get -v github.com/akiyosi/qt/cmd/... && $(go env GOPATH)/bin/qtsetup test && $(go env GOPATH)/bin/qtsetup -test=false
