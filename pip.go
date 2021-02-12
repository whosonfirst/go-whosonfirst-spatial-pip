package pip

import (
	"context"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-sfomuseum-mapshaper"
	"github.com/skelterjohn/geom"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
	"strconv"
)

type FilterSPRResultsFunc func(context.Context, reader.Reader, *geojson.Feature, []spr.StandardPlacesResult) (spr.StandardPlacesResult, error)

func SingleSPRResultsFunc(ctx context.Context, r reader.Reader, f *geojson.Feature, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	if len(possible) != 1 {
		return nil, fmt.Errorf("Number of results != 1")
	}

	parent_spr := possible[0]
	return parent_spr, nil
}

type Tool struct {
	Reader    reader.Reader
	Database  database.SpatialDatabase
	Mapshaper *mapshaper.Client
}

func NewPointInPolgygonTool(ctx context.Context, spatial_db database.SpatialDatabase, spatial_reader reader.Reader, ms_client *mapshaper.Client) (*Tool, error) {

	t := &Tool{
		Reader:    spatial_reader,
		Database:  spatial_db,
		Mapshaper: ms_client,
	}

	return t, nil
}

func (t *Tool) PointInPolygonAndUpdate(ctx context.Context, f *geojson.Feature, results_cb FilterSPRResultsFunc) ([]byte, error) {

	possible, err := t.PointInPolygon(ctx, f)

	if err != nil {
		return nil, err
	}

	parent_spr, err := results_cb(ctx, t.Reader, f, possible)

	if err != nil {
		return nil, err
	}

	parent_id, err := strconv.ParseInt(parent_spr.Id(), 10, 64)

	if err != nil {
		return nil, err
	}

	parent_f, err := wof_reader.LoadFeatureFromID(ctx, t.Reader, parent_id)

	if err != nil {
		return nil, err
	}

	body, err := f.MarshalJSON()

	if err != nil {
		return nil, err
	}

	parent_hierarchy := whosonfirst.Hierarchies(parent_f)
	parent_country := whosonfirst.Country(parent_f)

	to_update := map[string]interface{}{
		"properties.wof:parent_id": parent_id,
		"properties.wof:country":   parent_country,
		"properties.wof:hierarchy": parent_hierarchy,
	}

	for path, v := range to_update {

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

func (t *Tool) PointInPolygon(ctx context.Context, f *geojson.Feature) ([]spr.StandardPlacesResult, error) {

	props := f.Properties
	v, ok := props["wof:placetype"]

	if !ok {
		return nil, fmt.Errorf("Missing 'wof:placetype' property")
	}

	// Get the list of valid ancestors for public galleries (enclosure)

	pt, err := placetypes.GetPlacetypeByName(v.(string))

	if err != nil {
		return nil, fmt.Errorf("Failed to create new placetype for 'venue', %v", err)
	}

	roles := []string{
		"common",
		"optional",
		"common_optional",
	}

	ancestors := placetypes.AncestorsForRoles(pt, roles)

	centroid, err := t.PointInPolygonCentroid(ctx, f)

	if err != nil {
		return nil, err
	}

	lon := centroid.X()
	lat := centroid.Y()

	// Start PIP-ing the list of ancestors - stop at the first match

	possible := make([]spr.StandardPlacesResult, 0)

	for _, a := range ancestors {

		coord := &geom.Coord{
			X: lon,
			Y: lat,
		}

		// Ensure placetype and is current filters for ancestor

		inputs := &filter.SPRInputs{
			Placetypes: []string{a.Name},
			IsCurrent:  []string{"1"},
		}

		spr_filter, err := filter.NewSPRFilterFromInputs(inputs)

		if err != nil {
			return nil, fmt.Errorf("Failed to create SPR filter from input, %v", err)
		}

		rsp, err := t.Database.PointInPolygon(ctx, coord, spr_filter)

		if err != nil {
			return nil, fmt.Errorf("Failed to point in polygon for %v, %v", coord, err)
		}

		results := rsp.Results()

		if len(results) == 0 {
			continue
		}

		possible = results
		break
	}

	return possible, nil
}

func (t *Tool) PointInPolygonCentroid(ctx context.Context, f *geojson.Feature) (*orb.Point, error) {

	var candidate *geojson.Feature

	geojson_type := f.Geometry.GeoJSONType()

	switch geojson_type {
	case "Point":
		candidate = f
	case "MultiPoint":

		// not at all clear this is the best way to deal with things
		// (20210204/thisisaaronland)

		bound := f.Geometry.Bound()
		pt := bound.Center()

		candidate = geojson.NewFeature(pt)

	case "Polygon", "MultiPolygon":

		// this is not great but it's also not hard and making
		// the "perfect" mapshaper interface is yak-shaving right
		// now (20210204/thisisaaronland)

		fc := geojson.NewFeatureCollection()
		fc.Append(f)

		fc, err := t.Mapshaper.AppendCentroids(ctx, fc)

		if err != nil {
			return nil, fmt.Errorf("Failed to append centroids, %v", err)
		}

		f = fc.Features[0]

		candidate = geojson.NewFeature(f.Geometry)

		lat, lat_ok := f.Properties["mps:latitude"]
		lon, lon_ok := f.Properties["mps:longitude"]

		if lat_ok && lon_ok {

			pt := orb.Point{
				lat.(float64),
				lon.(float64),
			}

			candidate = geojson.NewFeature(pt)
		}

	default:
		return nil, fmt.Errorf("Unsupported type '%s'", t)
	}

	pt := candidate.Geometry.(orb.Point)
	return &pt, nil
}
