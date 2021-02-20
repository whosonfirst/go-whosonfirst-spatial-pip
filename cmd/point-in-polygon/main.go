package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "github.com/whosonfirst/go-writer-featurecollection"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/application"
	"log"
)

func main() {

	ctx := context.Background()

	opts, paths, err := application.NewApplicationOptionsFromCommandLine(ctx)

	if err != nil {
		log.Fatalf("Failed to create new PIP application opts, %v", err)
	}

	app, err := application.NewApplication(ctx, opts)

	if err != nil {
		log.Fatalf("Failed to create new PIP application, %v", err)
	}

	err = app.Run(ctx, paths)

	if err != nil {
		log.Fatalf("Failed to run PIP application, %v", err)
	}

}
