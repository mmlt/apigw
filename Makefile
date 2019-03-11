#.PHONY all build bake push login

VERSION?=0.0.7

MODULE=github.com/mmlt/apigw
MODULE_DIR=${GOPATH}/src/$(MODULE)

all: build bake push

build: 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" $(MODULE)

lint:
	gometalinter.v2 --vendor src/$(MODULE)/...

test: testu teste2e testresults

testu:
	cd $(MODULE_DIR) && ${GOPATH}/test.sh

teste2e:
	# -race
	cd $(MODULE_DIR) && go test -covermode=count -coverpkg="$(MODULE)/..." -coverprofile=teste2e.cov apigw_test.go

testresults:
	cd $(MODULE_DIR) && gocovmerge teste2e.cov test.cov > all.cov
	cd $(MODULE_DIR) && go tool cover -html all.cov
	cd $(MODULE_DIR) && go tool cover -func=all.cov