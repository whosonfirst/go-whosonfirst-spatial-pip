package query

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-server"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/api"
	"github.com/whosonfirst/go-whosonfirst-spatial/app"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"log"
	gohttp "net/http"
)

type QueryApplication struct {
}

func NewQueryApplication(ctx context.Context) (*QueryApplication, error) {

	query_app := &QueryApplication{}
	return query_app, nil
}

func (query_app *QueryApplication) RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "PIP")

	if err != nil {
		return err
	}

	err = flags.ValidateCommonFlags(fs)

	if err != nil {
		return err
	}

	err = flags.ValidateQueryFlags(fs)

	if err != nil {
		return err
	}

	err = flags.ValidateIndexingFlags(fs)

	if err != nil {
		return err
	}

	mode, err := lookup.StringVar(fs, "mode")

	if err != nil {
		return err
	}

	server_uri, err := lookup.StringVar(fs, "server-uri")

	if err != nil {
		return err
	}

	enable_geojson, err := lookup.BoolVar(fs, "enable-geojson")

	if err != nil {
		return err
	}

	spatial_app, err := app.NewSpatialApplicationWithFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to create new spatial application, %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	uris := fs.Args()

	go func() {

		err := spatial_app.Iterator.IterateURIs(ctx, uris...)

		if err != nil {
			log.Printf("Failed to iterate URIs, %v", err)
		}
	}()

	switch mode {

	case "cli":

		req, err := pip.NewPointInPolygonRequestFromFlagSet(fs)

		if err != nil {
			return fmt.Errorf("Failed to create SPR filter, %v", err)
		}

		rsp, err := pip.QueryPointInPolygon(ctx, spatial_app, req)

		if err != nil {
			return fmt.Errorf("Failed to query, %v", err)
		}

		enc, err := json.Marshal(rsp)

		if err != nil {
			return fmt.Errorf("Failed to marshal results, %v", err)
		}

		fmt.Println(string(enc))

	case "lambda":

		handler := func(ctx context.Context, req *pip.PointInPolygonRequest) (interface{}, error) {
			return pip.QueryPointInPolygon(ctx, spatial_app, req)
		}

		lambda.Start(handler)

	case "server":

		pip_opts := &api.PointInPolygonHandlerOptions{
			EnableGeoJSON: enable_geojson,
		}

		pip_handler, err := api.PointInPolygonHandler(spatial_app, pip_opts)

		if err != nil {
			return fmt.Errorf("Failed to create PIP handler, %v", err)
		}

		mux := gohttp.NewServeMux()
		mux.Handle("/", pip_handler)

		s, err := server.NewServer(ctx, server_uri)

		if err != nil {
			return fmt.Errorf("Failed to create server for '%s', %v", server_uri, err)
		}

		log.Printf("Listening for requests at %s\n", s.Address())

		err = s.ListenAndServe(ctx, mux)

		if err != nil {
			return fmt.Errorf("Failed to serve requests for '%s', %v", server_uri, err)
		}

	default:
		return fmt.Errorf("Invalid or unsupported mode '%s'", mode)
	}

	return nil
}
