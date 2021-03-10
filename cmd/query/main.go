package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/query"
	"log"
)

func main() {

	ctx := context.Background()

	fs, err := query.NewQueryApplicationFlagSet(ctx)

	if err != nil {
		log.Fatalf("Failed to create application flag set, %v", err)
	}

	app, err := query.NewQueryApplication(ctx)

	if err != nil {
		log.Fatalf("Failed to create new PIP application, %v", err)
	}

	err = app.RunWithFlagSet(ctx, fs)

	if err != nil {
		log.Fatalf("Failed to run PIP application, %v", err)
	}

}
