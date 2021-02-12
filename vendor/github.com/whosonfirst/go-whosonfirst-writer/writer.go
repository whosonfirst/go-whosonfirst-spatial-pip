package writer

import (
	"bytes"
	"context"
	"errors"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	go_writer "github.com/whosonfirst/go-writer"
	"io/ioutil"
)

func WriteFeature(ctx context.Context, wr go_writer.Writer, f geojson.Feature) error {
	return WriteFeatureBytes(ctx, wr, f.Bytes())
}

func WriteFeatureBytes(ctx context.Context, wr go_writer.Writer, body []byte) error {

	ex, err := export.NewExporter(ctx, "whosonfirst://")

	if err != nil {
		return err
	}
	
	ex_body, err := ex.Export(ctx, body)

	if err != nil {
		return err
	}
	
	id_rsp := gjson.GetBytes(ex_body, "properties.wof:id")

	if !id_rsp.Exists() {
		return errors.New("Missing 'properties.wof:id' property")
	}

	id := id_rsp.Int()

	rel_path, err := uri.Id2RelPath(id)

	if err != nil {
		return err
	}

	br := bytes.NewReader(ex_body)
	fh := ioutil.NopCloser(br)

	return wr.Write(ctx, rel_path, fh)
}
