package api

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/app"
	"net/http"
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
		}

		pip_rsp, err := pip.QueryPointInPolygon(ctx, app, pip_req)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
		}

		// geojson here?

		enc := json.NewEncoder(rsp)
		err = enc.Encode(pip_rsp)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	pip_handler := http.HandlerFunc(fn)
	return pip_handler, nil
}
