#!/usr/bin/env bash

go build -o thrifterc -v -ldflags "-X github.com/jxskiss/gothrifter/generator.GitRevision=`git rev-parse --short HEAD`"
