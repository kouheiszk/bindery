build:
	dep ensure -v
	go build -ldflags="-s -w" -o bin/bindery main.go log.go image.go temporary.go option.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock
