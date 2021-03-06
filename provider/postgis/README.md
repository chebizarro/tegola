# PostGIS
The PostGIS provider manages querying for tile requests against a Postgres database with the [PostGIS](http://postgis.net/) extension installed. The connection between tegola and Postgis is configured in a `tegola.toml` file. An example minimum connection config:


```toml
[[providers]]
name = "test_postgis"       # provider name is referenced from map layers (required)
type = "postgis"            # the type of data provider must be "postgis" for this data provider (required)
host = "localhost"          # PostGIS database host (required)
port = 5432                 # PostGIS database port (required)
database = "tegola"         # PostGIS database name (required)
user = "tegola"             # PostGIS database user (required)
password = ""               # PostGIS database password (required)
```

### Connection Properties

- `name` (string): [Required] provider name is referenced from map layers
- `type` (string): [Required] the type of data provider. must be "postgis" to use this data provider
- `host` (string): [Required] PostGIS database host
- `port` (int): [Required] PostGIS database port (required)
- `database` (string): [Required] PostGIS database name
- `user` (string): [Required] PostGIS database user
- `password` (string): [Required] PostGIS database password
- `srid` (int): [Optional] The default SRID for the provider. Defaults to WebMercator (3857) but also supports WGS84 (4326)
- `max_connections` (int): [Optional] The max connections to maintain in the connection pool. Defaults to 100. 0 means no max.

## Provider Layers
In addition to the connection configuration above, Provider Layers need to be configured. A Provider Layer tells tegola how to query PostGIS for a certain layer. An example minimum config:

```toml
[[providers.layers]]
name = "landuse"
# this table uses "geom" for the geometry_fieldname and "gid" for the id_fieldname so they don't need to be configured
tablename = "gis.zoning_base_3857"  
```

### Provider Layers Properties

- `name` (string): [Required] the name of the layer. This is used to reference this layer from map layers.
- `tablename` (string): [*Required] the name of the database table to query against. Required if `sql` is not defined.
- `geometry_fieldname` (string): [Optional] the name of the filed which contains the geometry for the feature. defaults to `geom`
- `id_fieldname` (string): [Optional] the name of the feature id field. defaults to `gid`
- `fields` ([]string): [Optional] a list of fields to include alongside the feature. Can be used if `sql` is not defined.
- `srid` (int): [Optional] the SRID of the layer. Supports `3857` (WebMercator) or `4326` (WGS84).
- `sql` (string): [*Required] custom SQL to use use. Required if `tablename` is not defined. Supports the following tokens:
  - !BBOX! - [Required] will be replaced with the bounding box of the tile before the query is sent to the database.
  - !ZOOM! - [Optional] will be replaced with the "Z" (zoom) value of the requested tile.


`*Required`: either the `tablename` or `sql` must be defined, but not both.

**Example minimum custom SQL config**

```toml
[[providers.layers]]
name = "rivers"
# custom SQL to be used for this layer. Note: that the geometery field is wrapped
# in ST_AsBinary() and a !BBOX! token is supplied for querying the table with the tile bounds
sql = "SELECT gid, ST_AsBinary(geom) AS geom FROM gis.rivers WHERE geom && !BBOX!"
```

## Testing
Testing is designed to work against a live PostGIS database. To run the PostGIS tests, the following environment variables need to be set:

```bash
$ export RUN_POSTGIS_TESTS=yes
$ export PGHOST="localhost"
$ export PGPORT=5432
$ export PGDATABASE="tegola"
$ export PGUSER="postgres"
$ export PGPASSWORD=""
```
