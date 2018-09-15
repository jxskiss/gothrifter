#!/usr/bin/env bash

packr build -o thriftkit -v -ldflags "-X github.com/jxskiss/thriftkit/generator.GitRevision=`git rev-parse --short HEAD`"
