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

_To be written_

### Iterators

_To be written_

### Exporters

_To be written_

### Writers

_To be written_

### Reverse geocoding centroids (and Mapshaper)

_To be written_

## Tools

```
$> make cli
go build -mod vendor -o bin/point-in-polygon cmd/point-in-polygon/main.go
```

### /point-in-polygon

Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.

```
> ./bin/point-in-polygon -h
Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.
Usage:
	 ./bin/point-in-polygon [options] uri(N) uri(N)
Valid options are:

  -exporter-uri string
    	A valid whosonfirst/go-whosonfirst-export URI. (default "whosonfirst://")
  -is-ceased value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-current value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-deprecated value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-superseded value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-superseding value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI scheme. This is used to identify WOF records to be PIP-ed. (default "repo://")
  -mapshaper-server string
    	A valid HTTP URI pointing to a sfomuseum/go-sfomuseum-mapshaper server endpoint. (default "http://localhost:8080")
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial URI. This is the database of spatial records that will for PIP-ing.
  -spatial-iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI scheme. This is used to identify WOF records to be indexed in the spatial database. (default "repo://")
  -spatial-source value
    	One or more URIs to be indexed in the spatial database (used for PIP-ing).
  -writer-uri string
    	A valid whosonfirst/go-writer URI. This is where updated records will be written to. (default "null://")
```

For example:

```
> ./bin/point-in-polygon \
	-writer-uri 'featurecollection://?writer=stdout://' \
	-spatial-database-uri 'sqlite://?dsn=:memory:' \
	-spatial-iterator-uri 'repo://?include=properties.mz:is_current=1' \
	-spatial-source /usr/local/data/sfomuseum-data-architecture \
	-iterator-uri 'repo://?include=properties.mz:is_current=1' \
	/usr/local/data/sfomuseum-data-publicart \
| jq '.features[]["properties"]["wof:parent_id"]' \
| sort \
| uniq \

-1
1159162825
1159162827
1477855979
1477855987
1477856005
1729791967
1729792389
1729792391
1729792433
1729792437
1729792459
1729792483
1729792489
1729792551
1729792577
1729792581
1729792643
1729792645
1729792679
1729792685
1729792689
1729792693
1729792699
```

So what's going on here?

The first thing we're saying is: Write features that have been PIP-ed to the [featurecollection](https://github.com/whosonfirst/go-writer-featurecollection) writer (which in turn in writing it's output to [STDOUT](https://github.com/whosonfirst/go-writer).

```
	-writer-uri 'featurecollection://?writer=stdout://'
```

Then we're saying: Create a new in-memory [SQLite spatial database](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite) to use for performing PIP operations.

```
	-spatial-database-uri 'sqlite://?dsn=:memory:' 
```

We're also going to create this spatial database on-the-fly by reading records in the `sfomuseum-data-architecture` respository selecting only records with a `mz:is_current=1` property.

```
	-spatial-iterator-uri 'repo://?include=properties.mz:is_current=1'
	-spatial-source /usr/local/data/sfomuseum-data-architecture 
```

If we already had a pre-built [SQLite database](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite#databases) we could specify it like this:

```
	-spatial-database-uri 'sqlite://?dsn=/path/to/sqlite.db' 
```

Next we define our _input_ data. This is the data is going to be PIP-ed. We are going to read records from the `sfomuseum-data-publicart` repository selecting only records with a `mz:is_current=1` property.

```
	-iterator-uri 'repo://?include=properties.mz:is_current=1' 
	/usr/local/data/sfomuseum-data-publicart 
```

Finally we pipe the results (a GeoJSON `FeatureCollection` string output to STDOUT) to the `jq` tool for filtering out `wof:parent_id` properties and then to the `sort` and `uniq` utlities to format the results.

```
| jq '.features[]["properties"]["wof:parent_id"]' 
| sort 
| uniq 
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-exporter
* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-writer
* https://github.com/sfomuseum/go-sfomuseum-mapshaper
