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
	"github.com/whosonfirst/go-whosonfirst-spatial/api"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spatial/properties"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"log"
	gohttp "net/http"
)

type QueryApplication struct {
}

func NewQueryApplication(ctx context.Context) (*QueryApplication, error) {

	app := &QueryApplication{}
	return app, nil
}

func (app *QueryApplication) RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

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

	database_uri, _ := lookup.StringVar(fs, "spatial-database-uri")
	properties_uri, _ := lookup.StringVar(fs, "properties-reader-uri")

	mode, err := lookup.StringVar(fs, "mode")

	if err != nil {
		return err
	}

	server_uri, err := lookup.StringVar(fs, "server-uri")

	if err != nil {
		return err
	}

	db, err := database.NewSpatialDatabase(ctx, database_uri)

	if err != nil {
		return err
	}

	query := func(ctx context.Context, req *api.PointInPolygonRequest) (interface{}, error) {

		c, err := geo.NewCoordinate(req.Longitude, req.Latitude)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new coordinate, %v", err)
		}

		f, err := api.NewSPRFilterFromPointInPolygonRequest(req)

		if err != nil {
			return nil, err
		}

		var rsp interface{}

		r, err := db.PointInPolygon(ctx, c, f)

		if err != nil {
			return nil, fmt.Errorf("Failed to query database with coord %v, %v", c, err)
		}

		rsp = r

		if len(req.Properties) > 0 {

			pr, err := properties.NewPropertiesReader(ctx, properties_uri)

			if err != nil {
				return nil, fmt.Errorf("Failed to create properties reader, %v", err)
			}

			r, err := pr.PropertiesResponseResultsWithStandardPlacesResults(ctx, rsp.(spr.StandardPlacesResults), req.Properties)

			if err != nil {
				return nil, fmt.Errorf("Failed to generate properties response, %v", err)
			}

			rsp = r
		}

		return r, nil
	}

	switch mode {

	case "cli":

		req, err := api.NewPointInPolygonRequestFromFlagSet(fs)

		if err != nil {
			return fmt.Errorf("Failed to create SPR filter, %v", err)
		}

		rsp, err := query(ctx, req)

		if err != nil {
			return fmt.Errorf("Failed to query, %v", err)
		}

		enc, err := json.Marshal(rsp)

		if err != nil {
			return fmt.Errorf("Failed to marshal results, %v", err)
		}

		fmt.Println(string(enc))

	case "lambda":

		lambda.Start(query)

	case "server":

		fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

			ctx := req.Context()

			var pip_req *api.PointInPolygonRequest

			dec := json.NewDecoder(req.Body)
			err := dec.Decode(&pip_req)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			}

			pip_rsp, err := query(ctx, pip_req)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			}

			enc := json.NewEncoder(rsp)
			err = enc.Encode(pip_rsp)

			if err != nil {
				gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			}

			return
		}

		pip_handler := gohttp.HandlerFunc(fn)

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
