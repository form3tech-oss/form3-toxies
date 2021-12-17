# form3-toxies

A custom [Toxiproxy](https://github.com/Shopify/toxiproxy) binary containing a few additional [custom toxics](https://github.com/Shopify/toxiproxy/blob/master/CREATING_TOXICS.md) (described below).

## Usage

### Docker

A docker image is available on [Docker Hub](https://hub.docker.com/r/form3tech/form3-toxies).

```
$ docker pull docker pull form3tech/form3-toxies
$ docker run --rm -it form3tech/form3-toxies`
```

The default Toxiproxy http API port is 8474. This will need to be mapped in order to configure Toxiproxy via the API.

## Custom Toxics

See the [Toxiproxy README](https://github.com/Shopify/toxiproxy) for details on configuring toxics.

### psql

Inject failure into a postgres connection, the failure can be triggered after a predefined number of
SQL statements have been sent to the server.

Attributes:

* `failure_type` - set to `ConnectionFailure` or `SyntaxError`. `ConnectionFailure` will terminate the connection, possibly 
causing retries depending on client implementation**. `SyntaxError` will inject a bad SQL statement into the connection 
causing a soft failure.
* `search_text` - regular expression used to match SQL statements
* `fail_on` - number of statements matching `search_text` to trigger failure on
* `recover_after` - number of statements matching `search_text` to trigger recovery after

** The golang psql client implementation will automatically retry on this type of failure. 