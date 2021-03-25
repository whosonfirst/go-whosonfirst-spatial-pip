package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-rtree"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/query"
	"log"
)

func main() {

	ctx := context.Background()

	app, err := query.NewQueryApplication(ctx)

	if err != nil {
		log.Fatalf("Failed to create new PIP application, %v", err)
	}

	err = app.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run PIP application, %v", err)
	}

}
