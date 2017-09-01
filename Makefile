_GOPATH 			:= $(PWD)/../../../..
export GOPATH := $(_GOPATH)

.PHONY: all
all: code

.PHONY: code
code:
	go build `go list ./... | grep -v /vendor/ `

.PHONY: test
test:
	go test -v `go list ./... | grep -v /vendor/ `
