#!/usr/bin/sh

go build -o thrifterc -v -ldflags "-X github.com/jxskiss/gothrifter/generator.Version=`git.exe rev-parse --short HEAD`"
