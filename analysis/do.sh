#!/bin/bash
sourceDir=$1
go build classifier.go
./classifier $sourceDir
