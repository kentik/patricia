_GOPATH 			:= $(PWD)/../../../..
export GOPATH := $(_GOPATH)
GENERATED_TYPES := bool string int int8 int16 int32 int64 uint uint8 uint16 uint32 uint64 uintptr byte rune float32 float64 complex64 complex128

.PHONY: all
all: codegen code

codegen: $(addprefix codegen-,$(GENERATED_TYPES))

codegen-%:
	@echo "** generating $* tree"
	mkdir -p "./${*}_tree"
	cp -pa template/*.go "./${*}_tree"
	rm -f ./${*}_tree/*_test.go
	rm -f ./${*}_tree/types.go
	( cd "${*}_tree" && sed -i "s/GeneratedType/${*}/g" *.go )
	( cd "${*}_tree" && sed -i "s/package template/package ${*}_tree/g" *.go )

.PHONY: clean
clean:
	rm -rf *_tree

.PHONY: code
code:
	go build `go list ./... | grep -v /vendor/ `

.PHONY: test
test:
	go test -v `go list ./... | grep -v /vendor/ `
