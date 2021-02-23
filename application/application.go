package application

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-sfomuseum-mapshaper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-iterate/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer"
	"github.com/whosonfirst/go-writer"
	"io"
	_ "log"
	"os"
)

type ApplicationOptions struct {
	Writer             writer.Writer
	WriterURI          string
	Exporter           export.Exporter
	ExporterURI        string
	MapshaperServerURI string
	SpatialDatabase    database.SpatialDatabase
	SpatialDatabaseURI string
	ToIterator         string
	FromIterator       string
		SPRFilterInputs    *filter.SPRInputs
	SPRResultsFunc     pip.FilterSPRResultsFunc		// This one chooses one result among many (or nil)
	PIPUpdateFunc pip.PointInPolygonToolUpdateCallback	// This one constructs a map[string]interface{} to update the target record (or not) 
}

type ApplicationPaths struct {
	To   []string
	From []string
}

type Application struct {
	to              string
	from            string
	tool            *pip.PointInPolygonTool
	writer          writer.Writer
	exporter        export.Exporter
	spatial_db      database.SpatialDatabase
	sprResultsFunc  pip.FilterSPRResultsFunc
	sprFilterInputs *filter.SPRInputs
	pipUpdateFunc pip.PointInPolygonToolUpdateCallback	
}

func NewApplicationOptionsFromCommandLine(ctx context.Context) (*ApplicationOptions, *ApplicationPaths, error) {

	iterator_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI scheme. This is used to identify WOF records to be PIP-ed.")

	exporter_uri := flag.String("exporter-uri", "whosonfirst://", "A valid whosonfirst/go-whosonfirst-export URI.")
	writer_uri := flag.String("writer-uri", "null://", "A valid whosonfirst/go-writer URI. This is where updated records will be written to.")

	spatial_database_uri := flag.String("spatial-database-uri", "", "A valid whosonfirst/go-whosonfirst-spatial URI. This is the database of spatial records that will for PIP-ing.")
	spatial_iterator_uri := flag.String("spatial-iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI scheme. This is used to identify WOF records to be indexed in the spatial database.")

	var spatial_paths multi.MultiString
	flag.Var(&spatial_paths, "spatial-source", "One or more URIs to be indexed in the spatial database (used for PIP-ing).")

	// As in github:sfomuseum/go-sfomuseum-mapshaper and github:sfomuseum/docker-sfomuseum-mapshaper
	// One day the functionality exposed here will be ported to Go and this won't be necessary

	mapshaper_server := flag.String("mapshaper-server", "http://localhost:8080", "A valid HTTP URI pointing to a sfomuseum/go-sfomuseum-mapshaper server endpoint.")

	var is_current multi.MultiString
	flag.Var(&is_current, "is-current", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	var is_ceased multi.MultiString
	flag.Var(&is_ceased, "is-ceased", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	var is_deprecated multi.MultiString
	flag.Var(&is_deprecated, "is-deprecated", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	var is_superseded multi.MultiString
	flag.Var(&is_superseded, "is-superseded", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	var is_superseding multi.MultiString
	flag.Var(&is_superseding, "is-superseding", "One or more existential flags (-1, 0, 1) to filter PIP results.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	inputs := &filter.SPRInputs{}

	inputs.IsCurrent = is_current
	inputs.IsCeased = is_ceased
	inputs.IsDeprecated = is_deprecated
	inputs.IsSuperseded = is_superseded
	inputs.IsSuperseding = is_superseding

	opts := &ApplicationOptions{
		WriterURI:          *writer_uri,
		ExporterURI:        *exporter_uri,
		SpatialDatabaseURI: *spatial_database_uri,
		MapshaperServerURI: *mapshaper_server,
		SPRResultsFunc:     pip.FirstButForgivingSPRResultsFunc, // sudo make me configurable
		SPRFilterInputs:    inputs,
		ToIterator:         *iterator_uri,
		FromIterator:       *spatial_iterator_uri,
	}

	pip_paths := flag.Args()

	paths := &ApplicationPaths{
		To:   pip_paths,
		From: spatial_paths,
	}

	return opts, paths, nil
}

func NewApplication(ctx context.Context, opts *ApplicationOptions) (*Application, error) {

	var ex export.Exporter
	var wr writer.Writer
	var spatial_db database.SpatialDatabase

	if opts.Exporter != nil {
		ex = opts.Exporter
	} else {

		_ex, err := export.NewExporter(ctx, opts.ExporterURI)

		if err != nil {
			return nil, fmt.Errorf("Failed to create exporter for '%s', %v", opts.ExporterURI, err)
		}

		ex = _ex
	}

	if opts.Writer != nil {
		wr = opts.Writer
	} else {
		_wr, err := writer.NewWriter(ctx, opts.WriterURI)

		if err != nil {
			return nil, fmt.Errorf("Failed to create writer for '%s', %v", opts.WriterURI, err)
		}

		wr = _wr
	}

	if opts.SpatialDatabase != nil {
		spatial_db = opts.SpatialDatabase
	} else {
		_db, err := database.NewSpatialDatabase(ctx, opts.SpatialDatabaseURI)

		if err != nil {
			return nil, fmt.Errorf("Failed to create spatial database for '%s', %v", opts.SpatialDatabaseURI, err)
		}

		spatial_db = _db
	}

	// All of this mapshaper stuff can't be retired/replaced fast enough...
	// (20210222/thisisaaronland)

	var ms_client *mapshaper.Client

	if opts.MapshaperServerURI != "" {

		// Set up mapshaper endpoint (for deriving centroids during PIP operations)
		// Make sure it's working

		client, err := mapshaper.NewClient(ctx, opts.MapshaperServerURI)

		if err != nil {
			return nil, fmt.Errorf("Failed to create mapshaper client for '%s', %v", opts.MapshaperServerURI, err)
		}

		ok, err := client.Ping()

		if err != nil {
			return nil, fmt.Errorf("Failed to ping '%s', %v", opts.MapshaperServerURI, err)
		}

		if !ok {
			return nil, fmt.Errorf("'%s' returned false", opts.MapshaperServerURI)
		}

		ms_client = client
	}

	update_cb := opts.PIPUpdateFunc
	
	if update_cb == nil {

		update_cb = pip.DefaultPointInPolygonToolUpdateCallback(spatial_db)
	}
	
	tool, err := pip.NewPointInPolygonTool(ctx, spatial_db, ms_client)

	if err != nil {
		return nil, fmt.Errorf("Failed to create PIP tool, %v", err)
	}

	app := &Application{
		to:              opts.ToIterator,
		from:            opts.FromIterator,
		spatial_db:      spatial_db,
		tool:            tool,
		exporter:        ex,
		writer:          wr,
		sprFilterInputs: opts.SPRFilterInputs,
		sprResultsFunc:  opts.SPRResultsFunc,
		pipUpdateFunc: update_cb,
	}

	return app, nil
}

func (app *Application) Run(ctx context.Context, paths *ApplicationPaths) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// These are the data we are indexing to PIP from

	err := app.IndexSpatialDatabase(ctx, paths.From...)

	if err != nil {
		return err
	}

	// These are the data we are PIP-ing

	to_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		path, err := emitter.PathForContext(ctx)

		if err != nil {
			return err
		}

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %v", path, err)
		}

		_, err = app.UpdateAndPublishFeature(ctx, body)

		if err != nil {
			return fmt.Errorf("Failed to update feature for '%s', %v", path, err)
		}

		return nil
	}

	to_iter, err := iterator.NewIterator(ctx, app.to, to_cb)

	if err != nil {
		return fmt.Errorf("Failed to create new PIP (to) iterator for input, %v", err)
	}

	err = to_iter.IterateURIs(ctx, paths.To...)

	if err != nil {
		return err
	}

	// This is important for something things like
	// whosonfirst/go-writer-featurecollection
	// (20210219/thisisaaronland)

	return app.writer.Close(ctx)
}

func (app *Application) IndexSpatialDatabase(ctx context.Context, uris ...string) error {

	from_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		switch geometry.Type(f) {
		case "Polygon", "MultiPolygon":
			return app.spatial_db.IndexFeature(ctx, f)
		default:
			return nil
		}
	}

	from_iter, err := iterator.NewIterator(ctx, app.from, from_cb)

	if err != nil {
		return fmt.Errorf("Failed to create spatial (from) iterator, %v", err)
	}

	err = from_iter.IterateURIs(ctx, uris...)

	if err != nil {
		return err
	}

	return nil
}

func (app *Application) UpdateAndPublishFeature(ctx context.Context, body []byte) ([]byte, error) {

	new_body, err := app.UpdateFeature(ctx, body)

	if err != nil {
		return nil, err
	}

	return app.PublishFeature(ctx, new_body)
}

func (app *Application) UpdateFeature(ctx context.Context, body []byte) ([]byte, error) {

	return app.tool.PointInPolygonAndUpdate(ctx, app.sprFilterInputs, app.sprResultsFunc, app.pipUpdateFunc, body)
}

func (app *Application) PublishFeature(ctx context.Context, body []byte) ([]byte, error) {

	new_body, err := app.exporter.Export(ctx, body)

	if err != nil {
		return nil, err
	}

	err = wof_writer.WriteFeatureBytes(ctx, app.writer, new_body)

	if err != nil {
		return nil, err
	}

	return new_body, nil
}

// UNTESTED

func (app *Application) WranglePIP(ctx context.Context, old_body []byte, pip_body []byte) ([]byte, error) {

	pip_parent_rsp := gjson.GetBytes(pip_body, "properties.wof:parent_id")

	if !pip_parent_rsp.Exists() {
		return nil, fmt.Errorf("Missing wof:parent_id")
	}

	pip_parent_id := pip_parent_rsp.Int()

	old_parent_rsp := gjson.GetBytes(old_body, "properties.wof:parent_id")

	if !old_parent_rsp.Exists() {
		return nil, fmt.Errorf("Missing wof:parent_id")
	}

	old_parent_id := old_parent_rsp.Int()

	if old_parent_id == pip_parent_id {
		return nil, nil
	}

	// Parent ID is different so we clone the old record, update the new record
	// as necesasary and supersede the former with the latter

	var new_body []byte
	var new_id int64

	old_id_rsp := gjson.GetBytes(old_body, "properties.wof:id")

	if !old_id_rsp.Exists() {
		return nil, fmt.Errorf("Missing wof:id")
	}

	old_id := old_id_rsp.Int()

	pip_parent_f, err := wof_reader.LoadFeatureFromID(ctx, app.spatial_db, pip_parent_id)

	if err != nil {
		return nil, err
	}

	pip_parent_inception := whosonfirst.Inception(pip_parent_f)
	pip_parent_hierarchy := whosonfirst.Hierarchies(pip_parent_f)

	new_body = old_body

	new_body, err = sjson.DeleteBytes(new_body, "properties.wof:id")

	if err != nil {
		return nil, err
	}

	new_body, err = app.exporter.Export(ctx, new_body)

	if err != nil {
		return nil, err
	}

	id_rsp := gjson.GetBytes(new_body, "properties.wof:id")

	if !id_rsp.Exists() {
		return nil, fmt.Errorf("Missing wof:id")
	}

	new_id = id_rsp.Int()

	old_updates := map[string]interface{}{
		"properties.mz:is_current":     0,
		"properties.wof:superseded_by": []int64{new_id},
		"properties.edtf:cessation":    pip_parent_inception,
	}

	for k, v := range old_updates {

		old_body, err = sjson.SetBytes(old_body, k, v)

		if err != nil {
			return nil, err
		}
	}

	new_updates := map[string]interface{}{
		"properties.wof:parent_id":  pip_parent_id,
		"properties.wof:hierarchy":  pip_parent_hierarchy,
		"properties.mz:is_current":  1,
		"properties.edtf:inception": pip_parent_inception,
		"properties.edtf:cessation": edtf.OPEN,
		"properties.wof:supersedes": []int64{old_id},
	}

	for k, v := range new_updates {

		new_body, err = sjson.SetBytes(new_body, k, v)

		if err != nil {
			return nil, err
		}
	}

	return new_body, nil
}
