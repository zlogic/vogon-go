# Vogon

[![Build status](https://github.com/zlogic/vogon-go/actions/workflows/build-go.yml/badge.svg?branch=master)](https://github.com/zlogic/vogon-go/actions)

Vogon is a simple web-based personal finance tracker written in Go.

Development is in progress, and the project is not production-ready or useful in any way.

## Project description

Simple web-based personal finance tracker using

* Plain Javascript and Bulma CSS on client-side
* Go on the server side
* [Pogreb](https://github.com/akrylysov/pogreb) key-value store for data storage

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

To disable request logging, set the `LOG_REQUESTS` environment variable to `false`.

## How to run the Docker image

To create a Vogon container, run the following Docker command (replace UID and port if necessary):
```
# UID for the process
VOGON_UID=10001
# Listen port
LISTEN_PORT=8080
# Create volume and fix permissions
docker volume create vogon-go
docker run --rm -v vogon-go:/data/vogon alpine chown $VOGON_UID:0 /data/vogon
# Create the container
docker create \
  --env DATABASE_DIR=/data/vogon \
  --env ALLOW_REGISTRATION=true \
  --publish $LISTEN_PORT:8080 \
  --user $VOGON_UID \
  -v vogon-go:/data/vogon:Z \
  ghcr.io/zlogic/vogon-go:latest
```

This will create a container with an embedded data store and allow registration.

Configuration of Vogon can be done via environment variables, as described in the [Configuration](#configuration) section above.

Supply the configuration via `--env` properties when creating the Docker container.

## How to run in standalone mode

Build Vogon:

`go build`

Run tests:

`go test ./...`

Run Vogon:

`vogon-go`

# Other versions

Vogon was previously using [Badger](https://github.com/dgraph-io/badger) DB for storing data.
That version is available in the [Badger branch](../../tree/badger).
