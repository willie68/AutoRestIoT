@echo off
go build -ldflags="-s -w" -o autorest-srv.exe cmd/service.go