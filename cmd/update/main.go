package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	_ "github.com/whosonfirst/go-writer-featurecollection"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/application/update"
	"log"
)

func main() {

	ctx := context.Background()

	fs, err := update.NewUpdateApplicationFlagSet(ctx)

	if err != nil {
		log.Fatalf("Failed to create application flag set, %v", err)
	}

	opts, paths, err := update.NewUpdateApplicationOptionsFromFlagSet(ctx, fs)

	if err != nil {
		log.Fatalf("Failed to create new PIP application opts, %v", err)
	}

	app, err := update.NewUpdateApplication(ctx, opts)

	if err != nil {
		log.Fatalf("Failed to create new PIP application, %v", err)
	}

	err = app.Run(ctx, paths)

	if err != nil {
		log.Fatalf("Failed to run PIP application, %v", err)
	}

}
