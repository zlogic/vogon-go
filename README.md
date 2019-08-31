# Vogon
Vogon is a simple web-based personal finance tracker written in Go.

Development is in progress, and the project is not production-ready or useful in any way.

## Project description

Simple web-based personal finance tracker using

* jQuery and Bootstrap 4 on client-side
* Go on the server side
* Badger key-value DB for data storage

Named after the Vogons (http://en.wikipedia.org/wiki/Vogon) race who were known to be extremely boring accountants.

_A rewrite from the [Node.js version of Vogon](https://github.com/zlogic/vogon-nj)_.

## Environment support

Vogon should work on most operating system supporting Go.
No external dependencies (such as a database) are required.

## Configuration

Set the `DATABASE_DIR` variable to the path where the database should be stored (e.g. `/data/vogon`).
Since all data will be stored in that directory, it's critical to keep it across restarts.

If you do not want random people using your deployment, you may want to set the `ALLOW_REGISTRATION` environment variable to `false`.

You should set `ALLOW_REGISTRATION` to `false` only after registering yourself.

## How to run the Docker image

To build the Docker image for Vogon container, run the following Docker command:

`docker build -t vogon:latest .`

To deploy create a Vogon container, run the following Docker command (change `[port]` to the port where Vogon-NJ will be accessible):

```
docker create \
  --env DATABASE_DIR=/data/vogon \
  --env ALLOW_REGISTRATION=true \
  --publish [port]:8080 \
  vogon:latest
```

This will create a container with an embedded Badger DB and allow registration.

Configuration of Vogon can be done via environment variables, as described in the [Configuration](#configuration) section above.

Supply the configuration via `--env` properties when creating the Docker container.

## How to run in standalone mode

Build Vogon:

`go build`

Run tests:

`go test ./...`

Run Vogon:

`vogon-go`
