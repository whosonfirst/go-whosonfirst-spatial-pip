cli:
	go build -mod vendor -o bin/query cmd/query/main.go
	go build -mod vendor -o bin/update cmd/update/main.go
