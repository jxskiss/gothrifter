#!/usr/bin/env bash

packr build -o thrifterc -v -ldflags "-X github.com/jxskiss/gothrifter/generator.GitRevision=`git rev-parse --short HEAD`"
