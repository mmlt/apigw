#.PHONY all build bake push login

VERSION?=v0.0.9
MODULE=github.com/mmlt/apigw


all: build test release

build: 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" $(MODULE)

lint:
	# TODO replace "gometalinter.v2 --vendor src/$(MODULE)/..." with golangci-lint 

test: testu teste2e testresults

testunit:
	./hack/test.sh

teste2e:
	# -race
	go test -covermode=count -coverpkg="$(MODULE)/..." -coverprofile=teste2e.cov apigw_test.go

testresults:
	gocovmerge teste2e.cov test.cov > all.cov
	go tool cover -html all.cov
	go tool cover -func=all.cov

release:
	./hack/release.sh $(VERSION)

