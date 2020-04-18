#!/usr/bin/env bash
GOOS=linux go build -o main main.go
zip deployment.zip main