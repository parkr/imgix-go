test: deps
	go test -cover

deps:
	go get github.com/stretchr/testify golang.org/x/tools/cmd/cover
