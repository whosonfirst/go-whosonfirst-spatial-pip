package query

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

const ENABLE_GEOJSON string = "enable-geojson"
const SERVER_URI string = "server-uri"
const MODE string = "mode"

func NewQueryApplicationFlagSet(ctx context.Context) (*flag.FlagSet, error) {

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

	fs.String(MODE, "cli", "...")
	fs.String(SERVER_URI, "http://localhost:8080", "...")
	fs.Bool(ENABLE_GEOJSON, false, "...")

	return fs, nil
}
