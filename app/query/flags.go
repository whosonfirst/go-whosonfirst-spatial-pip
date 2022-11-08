package query

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

const ENABLE_GEOJSON string = "enable-geojson"
const SERVER_URI string = "server-uri"
const MODE string = "mode"

var mode string
var server_uri string
var enable_geojson bool

func DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs, err := flags.CommonFlags()

	if err != nil {
		return nil, err
	}

	err = flags.AppendQueryFlags(fs)

	if err != nil {
		return nil, err
	}

	err = flags.AppendIndexingFlags(fs)

	if err != nil {
		return nil, err
	}

	fs.StringVar(&mode, "mode", "cli", "...")
	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "...")
	fs.BoolVar(&enable_geojson, "enable-geojson", false, "...")

	return fs, nil
}
