# go-whosonfirst-spatial-pip

## IMPORTANT

This is work in progress. Documentation to follow.

## Example

```
package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/application"
)

func main() {

	ctx := context.Background()

	opts, paths, _ := application.NewApplicationOptionsFromCommandLine(ctx)
	app, _ := application.NewApplication(ctx, opts)
	app.Run(ctx, paths)
}
```

_Error handling omitted for the sake of brevity._

## Concepts

### "Spatial" databases

### Iterators

### Exporters

### Writers

## Tools

```
$> make cli
go build -mod vendor -o bin/point-in-polygon cmd/point-in-polygon/main.go
```

### /point-in-polygon

For example:

```
> ./bin/point-in-polygon \
	-writer-uri stdout:// \
	-spatial-database-uri 'sqlite://?dsn=:memory:' \
	-spatial-iterator-uri 'repo://?include=properties.mz:is_current=1' \
	-spatial-source /usr/local/data/sfomuseum-data-architecture \
	-iterator-uri 'repo://?include=properties.mz:is_current=1' \
	/usr/local/data/sfomuseum-data-publicart/
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-exporter
* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-writer
* https://github.com/sfomuseum/go-sfomuseum-mapshaper
