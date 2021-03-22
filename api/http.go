package api

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/app"
	"github.com/whosonfirst/go-whosonfirst-spr-geojson"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"net/http"
	"github.com/aaronland/go-http-sanitize"
)

type PointInPolygonHandlerOptions struct {
	EnableGeoJSON bool
}

func PointInPolygonHandler(app *spatial_app.SpatialApplication, opts *PointInPolygonHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		var pip_req *pip.PointInPolygonRequest

		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&pip_req)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		pip_rsp, err := pip.QueryPointInPolygon(ctx, app, pip_req)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		format, err := sanitize.GetString(req, "format")

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		if opts.EnableGeoJSON && format == "geojson" {

			opts := &geojson.AsFeatureCollectionOptions{
				Reader: app.SpatialDatabase,
				Writer: rsp,
			}

			err := geojson.AsFeatureCollection(ctx, pip_rsp.(spr.StandardPlacesResults), opts)

			if err != nil {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		// geojson here?

		enc := json.NewEncoder(rsp)
		err = enc.Encode(pip_rsp)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	pip_handler := http.HandlerFunc(fn)
	return pip_handler, nil
}
