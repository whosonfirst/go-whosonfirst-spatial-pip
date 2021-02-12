package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-json-query"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-sfomuseum-mapshaper"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer"
	"github.com/whosonfirst/go-writer"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

func main() {

	indexer_uri := flag.String("indexer-uri", "repo://", "A valid whosonfirst/go-whosonfirst-index URI.")
	exporter_uri := flag.String("exporter-uri", "whosonfirst://", "A valid whosonfirst/go-whosonfirst-export URI.")
	writer_uri := flag.String("writer-uri", "null://", "A valid whosonfirst/go-writer URI.")

	spatial_uri := flag.String("spatial-database-uri", "", "A valid whosonfirst/go-whosonfirst-spatial URI.")
	spatial_mode := flag.String("spatial-mode", "repo://", "...")

	var spatial_paths multi.MultiString
	flag.Var(&spatial_paths, "spatial-source", "...")

	// As in github:sfomuseum/go-sfomuseum-mapshaper and github:sfomuseum/docker-sfomuseum-mapshaper
	// One day the functionality exposed here will be ported to Go and this won't be necessary

	mapshaper_server := flag.String("mapshaper-server", "http://localhost:8080", "A valid HTTP URI pointing to a sfomuseum/go-sfomuseum-mapshaper server endpoint.")

	var queries query.QueryFlags
	flag.Var(&queries, "query", "One or more {PATH}={REGEXP} parameters for filtering records.")

	valid_query_modes := strings.Join([]string{query.QUERYSET_MODE_ALL, query.QUERYSET_MODE_ANY}, ", ")
	desc_query_modes := fmt.Sprintf("Specify how query filtering should be evaluated. Valid modes are: %s", valid_query_modes)

	query_mode := flag.String("query-mode", query.QUERYSET_MODE_ALL, desc_query_modes)

	var is_current multi.MultiString
	flag.Var(&is_current, "is-current", "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_ceased multi.MultiString
	flag.Var(&is_ceased, "is-ceased", "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_deprecated multi.MultiString
	flag.Var(&is_deprecated, "is-deprecated", "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_superseded multi.MultiString
	flag.Var(&is_superseded, "is-superseded", "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_superseding multi.MultiString
	flag.Var(&is_superseding, "is-superseding", "One or more existential flags (-1, 0, 1) to filter results by.")
	
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ex, err := export.NewExporter(ctx, *exporter_uri)

	if err != nil {
		log.Fatalf("Failed to create exporter for '%s', %v", *exporter_uri, err)
	}

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create writer for '%s', %v", *writer_uri, err)
	}

	// Set up mapshaper endpoint (for deriving centroids during PIP operations)
	// Make sure it's working

	ms_client, err := mapshaper.NewClient(ctx, *mapshaper_server)

	if err != nil {
		log.Fatalf("Failed to create mapshaper server for '%s', %v", *mapshaper_server, err)
	}

	ok, err := ms_client.Ping()

	if err != nil {
		log.Fatalf("Failed to ping '%s', %v", *mapshaper_server, err)
	}

	if !ok {
		log.Fatalf("'%s' returned false", *mapshaper_server)
	}

	var qs *query.QuerySet

	if len(queries) > 0 {

		qs = &query.QuerySet{
			Queries: queries,
			Mode:    *query_mode,
		}
	}

	//

	spatial_db, err := database.NewSpatialDatabase(ctx, *spatial_uri)

	if err != nil {
		log.Fatalf("Failed to create spatial database, %v", err)
	}

	if len(spatial_paths) > 0 {

		spatial_cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

			f, err := feature.LoadFeatureFromReader(fh)

			if err != nil {
				return err
			}

			switch geometry.Type(f) {
			case "Polygon", "MultiPolygon":
				return spatial_db.IndexFeature(ctx, f)
			default:
				return nil
			}
		}

		spatial_idx, err := index.NewIndexer(*spatial_mode, spatial_cb)

		if err != nil {
			log.Fatalf("Failed to create spatial indexer, %v", err)
		}

		err = spatial_idx.Index(ctx, spatial_paths...)

		if err != nil {
			log.Fatalf("Failed to index spatial paths, %v", err)
		}
	}

	//

	tool, err := pip.NewPointInPolygonTool(ctx, spatial_db, ms_client)

	if err != nil {
		log.Fatalf("Failed to create PIP tool, %v", err)
	}

	inputs := &filter.SPRInputs{}

	inputs.IsCurrent = is_current
	inputs.IsCeased = is_ceased
	inputs.IsDeprecated = is_deprecated
	inputs.IsSuperseded = is_superseded
	inputs.IsSuperseding = is_superseding	
	
	pip_cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
		}

		if qs != nil {

			matches, err := query.Matches(ctx, qs, body)

			if err != nil {
				return err
			}

			if !matches {
				return nil
			}
		}

		new_body, err := tool.PointInPolygonAndUpdate(ctx, inputs, pip.FirstSPRResultsFunc, body)

		if err != nil {
			return err
		}

		new_body, err = ex.Export(ctx, new_body)

		if err != nil {
			return err
		}

		err = wof_writer.WriteFeatureBytes(ctx, wr, new_body)

		if err != nil {
			return err
		}

		log.Println("Update", path)
		return nil
	}

	pip_idx, err := index.NewIndexer(*indexer_uri, pip_cb)

	if err != nil {
		log.Fatal(err)
	}

	pip_paths := flag.Args()

	err = pip_idx.Index(ctx, pip_paths...)

	if err != nil {
		log.Fatal(err)
	}
}
