# go-whosonfirst-spatial-pip

## IMPORTANT

This is work in progress. Documentation to follow.

## Background

This package exports point-in-polygon (PIP) applications using the `whosonfirst/go-whosonfirst-spatial` interfaces.

The code in this package does not contain any specific implementation of those interfaces so when invoked on its own it won't work as expected.

The code in this package is designed to be imported by _other_ code that also loads the relevant packages that implement the `whosonfirst/go-whosonfirst-spatial` interfaces. For example, here is the `query` application using the `whosonfirst/go-whosonfirst-spatial-sqlite` package. This application is part of the [go-whosonfirst-spatial-pip-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-pip-sqlite) package:

```
package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/query"
)

func main() {

	ctx := context.Background()

	fs, _ := query.NewQueryApplicationFlagSet(ctx)
	app, _ := query.NewQueryApplication(ctx)

	app.RunWithFlagSet(ctx, fs)
}
```

The idea is that this package defines code to implement opinionated applications without specifying an underlying database implementation (interface).

As of this writing this package exports two "applications":

* The "Query" application performs a basic point-in-polygon (PIP) query, with optional "standard places response" (SPR) filters.

* The "Update" application accepts a series of Who's On First (WOF) records and attempts to assign a "parent" ID and hierarchy by performing one or more PIP operations for that record's centroid and potential ancestors (derived from its placetype). If successful the application also tries to "write" the updated feature to a target that implements the `whosonfirst/go-writer` interface.

Although there is a substantial amount of overlap, conceptually, between the two applications not all those similarities have been reconciled. These include:

* The "Update" application will, optionally, attempt to populate (or index) a spatial database when it starts. The "Query" application does not yet.

* The "Query" application is designed to run in a number of different "modes". These are: As a command line application; As a standalone HTTP server; As an AWS Lambda function. The "Update" application currently only runs as a command line application.

## Applications

_The examples shown here assume applications that have been built with the [whosonfirst/go-whosonfirst-spatial-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite)._

### Query

#### Command line

```
$> ./bin/query \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/arch.db' \
	-latitude 37.616951 \
	-longitude -122.383747 \
	-is-current 1

| jq

{
  "places": [
    {
      "wof:id": "1729792433",
      "wof:parent_id": "1729792389",
      "wof:name": "Terminal 2 Main Hall",
      "wof:country": "US",
      "wof:placetype": "concourse",
      "mz:latitude": 37.617044,
      "mz:longitude": -122.383533,
      "mz:min_latitude": 37.61556454299907,
      "mz:min_longitude": 37.617044,
      "mz:max_latitude": -122.3849539833859,
      "mz:max_longitude": -122.38296693570045,
      "mz:is_current": 1,
      "mz:is_deprecated": 0,
      "mz:is_ceased": 1,
      "mz:is_superseded": 0,
      "mz:is_superseding": 1,
      "wof:path": "172/979/243/3/1729792433.geojson",
      "wof:repo": "sfomuseum-data-architecture",
      "wof:lastmodified": 1612909946
    },
    {
      "wof:id": "1729792685",
      "wof:parent_id": "1729792389",
      "wof:name": "Terminal Two Arrivals",
      "wof:country": "XX",
      "wof:placetype": "concourse",
      "mz:latitude": 37.617036431454586,
      "mz:longitude": -122.38394076589181,
      "mz:min_latitude": 37.61603604049649,
      "mz:min_longitude": 37.617036431454586,
      "mz:max_latitude": -122.3848417563672,
      "mz:max_longitude": -122.38330449541728,
      "mz:is_current": 1,
      "mz:is_deprecated": 0,
      "mz:is_ceased": 1,
      "mz:is_superseded": 0,
      "mz:is_superseding": 0,
      "wof:path": "172/979/268/5/1729792685.geojson",
      "wof:repo": "sfomuseum-data-architecture",
      "wof:lastmodified": 1612910034
    }
  ]
}
```

#### Server

#### Lambda

### Update

Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.

```
$> ./bin/point-in-polygon -h
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

#### Command line

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
