@echo off
go build -ldflags="-s -w" -o test-mqtt.exe main.go