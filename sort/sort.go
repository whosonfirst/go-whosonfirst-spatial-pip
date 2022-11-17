package sort

// This should probably go in go-whosonfirst-spr but it will do for now

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type SortedStandardPlacesResults struct {
	spr.StandardPlacesResults
	results []spr.StandardPlacesResult
}

func (r *SortedStandardPlacesResults) Results() []spr.StandardPlacesResult {
	return r.results
}

type Sorter interface {
	Sort(context.Context, spr.StandardPlacesResults) (spr.StandardPlacesResults, error)
}
