#!/usr/bin/env bash
env GOOS=linux GOARCH=amd64 go build -o upload_file -v upload_file.go
