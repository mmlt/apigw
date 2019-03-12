#.PHONY all build bake push login

VERSION?=0.0.7

MODULE=github.com/mmlt/apigw
MODULE_DIR=${GOPATH}/src/$(MODULE)

all: build bake push

build: 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" $(MODULE)

lint:
	# TODO replace "gometalinter.v2 --vendor src/$(MODULE)/..." with golangci-lint 

test: testu teste2e testresults

testunit:
	./test.sh

teste2e:
	# -race
	go test -covermode=count -coverpkg="$(MODULE)/..." -coverprofile=teste2e.cov apigw_test.go

testresults:
	gocovmerge teste2e.cov test.cov > all.cov
	go tool cover -html all.cov
	go tool cover -func=all.cov