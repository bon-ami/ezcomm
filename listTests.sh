#!/bin/sh
echo go test -v -verbose=3 -run
go test -list .
echo go test -v -verbose=3 -run