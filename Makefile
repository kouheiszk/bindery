build:
	dep ensure -v
	go build -ldflags="-s -w" -o bin/bindery main.go log.go version.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock
