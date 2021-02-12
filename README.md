# go-whosonfirst-spatial-pip

## IMPORTANT

This is work in progress. Documentation to follow.

## Tools

### /point-in-polygon

```
> ./bin/point-in-polygon -h
Usage of ./bin/point-in-polygon:
  -exporter-uri string
    	A valid whosonfirst/go-whosonfirst-export URI. (default "whosonfirst://")
  -indexer-uri string
    	A valid whosonfirst/go-whosonfirst-index URI. (default "repo://")
  -is-ceased value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-current value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-deprecated value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-superseded value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-superseding value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -mapshaper-server string
    	A valid HTTP URI pointing to a sfomuseum/go-sfomuseum-mapshaper server endpoint. (default "http://localhost:8080")
  -query value
    	One or more {PATH}={REGEXP} parameters for filtering records.
  -query-mode string
    	Specify how query filtering should be evaluated. Valid modes are: ALL, ANY (default "ALL")
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial URI.
  -spatial-mode string
    	... (default "repo://")
  -spatial-source value
    	...
  -writer-uri string
    	A valid whosonfirst/go-writer URI. (default "null://")
```

For example:

```
$> ./bin/point-in-polygon \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/whosonfirst/go-whosonfirst-spatial-sqlite/arch.db' \
	-query 'properties.mz:is_current=1' \
	-query 'properties.sfomuseum:placetype=gallery' \
	-is-current 1 \	
	-writer-uri null:// \
	/usr/local/data/sfomuseum-data-architecture/
	
2021/02/12 12:09:32 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/566/3/1477855663.geojson
2021/02/12 12:09:33 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/566/1/1477855661.geojson
2021/02/12 12:09:33 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/582/1/1477855821.geojson
2021/02/12 12:09:33 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/582/3/1477855823.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/1/1477855941.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/593/9/1477855939.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/3/1477855943.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/583/1/1477855831.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/5/1477855945.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/9/1477855949.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/7/1477855947.geojson
2021/02/12 12:09:34 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/595/5/1477855955.geojson
... and so on
```

Or:

```
> ./bin/point-in-polygon \
	-spatial-database-uri 'sqlite://?dsn=:memory:' \
	-spatial-source /usr/local/data/sfomuseum-data-architecture/ \
	-query 'properties.mz:is_current=1' \
	-query 'properties.sfomuseum:placetype=gallery' \
	-is-current 1	
	-writer-uri null:// \
	/usr/local/data/sfomuseum-data-architecture/
	
2021/02/12 12:11:35 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/566/1/1477855661.geojson
2021/02/12 12:11:36 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/582/1/1477855821.geojson
2021/02/12 12:11:36 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/566/3/1477855663.geojson
2021/02/12 12:11:36 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/582/3/1477855823.geojson
2021/02/12 12:11:36 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/1/1477855941.geojson
2021/02/12 12:11:37 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/594/3/1477855943.geojson
2021/02/12 12:11:37 Update /usr/local/data/sfomuseum-data-architecture/data/147/785/593/9/1477855939.geojson
... and so on
```

#### Inline queries

_To be written_

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
