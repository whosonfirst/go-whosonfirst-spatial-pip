package query

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aaronland/go-http-server"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-spatial/api"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spatial/properties"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"log"
	gohttp "net/http"
)

type QueryApplication struct {
	spatial_db        database.SpatialDatabase
	properties_reader properties.PropertiesReader
	mode              string
	server_uri        string
}

type QueryApplicationOptions struct {
}

func NewQueryApplicationOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*QueryApplicationOptions, error) {

	fs, err := flags.CommonFlags()

	if err != nil {
		return nil, err
	}

	err = flags.AppendQueryFlags(fs)

	if err != nil {
		return nil, err
	}

	mode := fs.String("mode", "cli", "...")

	server_uri := fs.String("server-uri", "http://localhost:8080", "...")

	flagset.Parse(fs)

	err = flags.ValidateCommonFlags(fs)

	if err != nil {
		return nil, err
	}

	err = flags.ValidateQueryFlags(fs)

	if err != nil {
		return nil, err
	}

	database_uri, _ := flags.StringVar(fs, "spatial-database-uri")
	properties_uri, _ := flags.StringVar(fs, "properties-reader-uri")

	opts := QueryApplicationOptions{
		SpatialDatabase:  database_uri,
		PropertiesReader: properties_uri,
		ServerURI:        *server_uri,
		Mode:             *mode,
	}

	return opts, nil
}

func NewQueryApplication(ctx context.Context, opts *QueryApplicationOptions) (*QueryApplication, error) {

	db, err := database.NewSpatialDatabase(ctx, opts.SpatialDatabase)

	if err != nil {
		return nil, err
	}

	app := &QueryApplication{
		spatial_db: db,
		mode:       opts.Mode,
		server_uri: opts.ServerURI,
	}

	if opts.PropertiesReader != "" {

		pr, err := properties.NewPropertiesReader(ctx, properties_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create properties reader, %v", err)
		}

		app.properties_reader = pr
	}

	return app, nil
}

func (app *QueryApplication) Run(ctx context.Context) error {

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

		r, err := app.spatial_database.PointInPolygon(ctx, c, f)

		if err != nil {
			return nil, fmt.Errorf("Failed to query database with coord %v, %v", c, err)
		}

		rsp = r

		if len(req.Properties) > 0 {

			if app.properties_reader == nil {
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

	switch app.mode {

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

		s, err := server.NewServer(ctx, app.server_uri)

		if err != nil {
			return fmt.Errorf("Failed to create server for '%s', %v", *server_uri, err)
		}

		log.Printf("Listening for requests at %s\n", s.Address())

		err = s.ListenAndServe(ctx, mux)

		if err != nil {
			return fmt.Errorf("Failed to serve requests for '%s', %v", *server_uri, err)
		}

	default:
		return fmt.Errorf("Invalid or unsupported mode '%s'", *mode)
	}

	return nil
}
