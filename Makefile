cli:
	go build -mod vendor -o bin/query examples/query/main.go
	go build -mod vendor -o bin/update examples/update/main.go
