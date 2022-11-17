package sort

import (
	"context"
	_ "fmt"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"sort"
)

type NameSorter struct {
	Sorter
}

func NewNameSorter(ctx context.Context, uri string) (Sorter, error) {
	s := &NameSorter{}
	return s, nil
}

func (s *NameSorter) Sort(ctx context.Context, results spr.StandardPlacesResults) (spr.StandardPlacesResults, error) {

	lookup := make(map[string][]spr.StandardPlacesResult)

	for _, s := range results.Results() {

		_results, ok := lookup[s.Name()]

		if !ok {
			_results = make([]spr.StandardPlacesResult, 0)
		}

		_results = append(_results, s)
		lookup[s.Name()] = _results

	}

	names := make([]string, 0)

	for n, _ := range lookup {
		names = append(names, n)
	}

	sort.Strings(names)

	sorted := make([]spr.StandardPlacesResult, 0)

	for _, n := range names {

		for _, s := range lookup[n] {
			sorted = append(sorted, s)
		}
	}

	sorted_results := &SortedStandardPlacesResults{
		results: sorted,
	}

	return sorted_results, nil
}