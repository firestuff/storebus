default: tidy test

tidy:
	go mod tidy
	goimports -l -w {{justfile_directory()}}
	gofumpt -l -w {{justfile_directory()}}
	go fmt {{justfile_directory()}}/...

test:
	go vet {{justfile_directory()}}/...
	golangci-lint run {{justfile_directory()}}/...
	go test -cover -race -timeout=7s -parallel=10 {{justfile_directory()}}/...
