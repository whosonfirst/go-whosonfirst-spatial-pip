package pip

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-spr"
	// "github.com/whosonfirst/go-reader"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/skelterjohn/geom"
	"github.com/sfomuseum/go-sfomuseum-mapshaper"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
)

type FilterSPRResultsFunc func(context.Context, []spr.StandardPlacesResult) (spr.StandardPlacesResult, error)

func PointInPolygon(ctx context.Context, spatial_db database.SpatialDatabase, ms_client *mapshaper.Client, f *geojson.Feature) ([]spr.StandardPlacesResult, error) {

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

	centroid, err := PointInPolygonCentroid(ctx, ms_client, f)

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

		rsp, err := spatial_db.PointInPolygon(ctx, coord, spr_filter)

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

func PointInPolygonCentroid(ctx context.Context, ms_client *mapshaper.Client, f *geojson.Feature) (*orb.Point, error) {

	var candidate *geojson.Feature

	t := f.Geometry.GeoJSONType()

	switch t {
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

		fc, err := ms_client.AppendCentroids(ctx, fc)

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
