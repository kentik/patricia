_GOPATH 			:= $(PWD)/../../../..
export GOPATH := $(_GOPATH)
GENERATED_TYPES := bool string int int8 int16 int32 int64 uint uint8 uint16 uint32 uint64 uintptr byte rune float32 float64 complex64 complex128

.PHONY: all
all: codegen code

codegen: $(addprefix codegen-,$(GENERATED_TYPES))
	@# handle interface{} separately
	mkdir -p interface_tree
	cp -pa template/*.go ./interface_tree
	rm -f ./interface_tree/*_test.go
	rm -f ./interface_tree/types.go
	( cd interface_tree && sed -i "s/GeneratedType/interface{}/g" *.go )
	( cd interface_tree && sed -i 's/package template/package interface_tree/g' *.go )
	#
	mkdir -p interface_array_tree
	cp -pa template/*.go ./interface_array_tree
	rm -f ./interface_array_tree/*_test.go
	rm -f ./interface_array_tree/types.go
	( cd interface_array_tree && sed -i "s/GeneratedType/[]interface{}/g" *.go )
	( cd interface_array_tree && sed -i 's/package template/package interface_array_tree/g' *.go )

codegen-%:
	@echo "** generating $* tree"
	mkdir -p "./${*}_tree"
	cp -pa template/*.go "./${*}_tree"
	rm -f ./${*}_tree/*_test.go
	rm -f ./${*}_tree/types.go
	( cd "${*}_tree" && sed -i "s/GeneratedType/${*}/g" *.go )
	( cd "${*}_tree" && sed -i "s/package template/package ${*}_tree/g" *.go )
	#
	@echo "** generating []$* tree"
	mkdir -p "./${*}_array_tree"
	cp -pa template/*.go "./${*}_array_tree"
	rm -f ./${*}_array_tree/*_test.go
	rm -f ./${*}_array_tree/types.go
	( cd "${*}_array_tree" && sed -i "s/GeneratedType/[]${*}/g" *.go )
	( cd "${*}_array_tree" && sed -i "s/package template/package ${*}_array_tree/g" *.go )
	#
	@echo "** generating *$* tree"
	mkdir -p "./${*}_ptr_tree"
	cp -pa template/*.go "./${*}_ptr_tree"
	rm -f ./${*}_ptr_tree/*_test.go
	rm -f ./${*}_ptr_tree/types.go
	( cd "${*}_ptr_tree" && sed -i "s/GeneratedType/\*${*}/g" *.go )
	( cd "${*}_ptr_tree" && sed -i "s/package template/package ${*}_ptr_tree/g" *.go )

.PHONY: clean
clean:
	rm -rf *_tree

.PHONY: code
code:
	go build `go list ./... | grep -v /vendor/ `

.PHONY: test
test:
	go test -v `go list ./... | grep -v /vendor/ `
