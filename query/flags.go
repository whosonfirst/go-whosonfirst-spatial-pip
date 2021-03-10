package query

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

func NewQueryApplicationFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs, err := flags.CommonFlags()

	if err != nil {
		return nil, err
	}

	err = flags.AppendQueryFlags(fs)

	if err != nil {
		return nil, err
	}

	fs.String("mode", "cli", "...")

	fs.String("server-uri", "http://localhost:8080", "...")

	return fs, nil
}
