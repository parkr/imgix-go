test: deps fmt
	go test -cover

deps:
	go get github.com/stretchr/testify golang.org/x/tools/cmd/cover

fmt:
	go fmt
