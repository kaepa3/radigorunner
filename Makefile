
.PHONY: all help installdeps build build-4-docker vet test test-cover docker-build

build:
	go build 

pi:
	GOOS=linux GOARCH=arm GOARM=6 go build -o pirunner pirunner.go
