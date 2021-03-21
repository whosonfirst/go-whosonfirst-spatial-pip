package api

import (
	"encoding/json"
	gohttp "net/http"
)

type PointInPolygonHandlerOptions struct {
	EnableGeoJSON    bool
	EnableProperties bool
	GeoJSONReader    reader.Reader
	SPRPathResolver  geojson.SPRPathResolver
}

func PointInPolygonHandler(spatial_app *app.SpatialApplication, opts *PointInPolygonHandlerOptions) (http.Handler, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		ctx := req.Context()
		
		var pip_req *api.PointInPolygonRequest
		
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&pip_req)
		
		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
		}

		// TBD???
		// spatial_app.Query
		// query(spatial_app)
		
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
	return pip_handler
}
