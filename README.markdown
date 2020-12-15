pg_info makes it a bit easier to access the various `pg_stat_*` informational
tables in PostgreSQL for getting metrics and such.

**This is really unfinished and very much work-in-progress. You almost certainly
don't want to use this yet.**

Usage
-----

<!--
There are some binaries available on the [release page][r]; these are statically
linked and don't have any dependencies.
-->

Compile from source with `go get zgo.at/pg_info/cmd/pg_info`; this will put the
binary in `~/go/bin`.

There are three ways to use this: as a web interface, a CLI program, or by
plugging it in to an existing Go web app:

### Run HTTP server

    $ pg_info serve -db '...'

The `-db` flag:

    -db "dbname=goatcounter_paths sslmode=disable"
    -db "postgres://host/db"

See `pg_info help serve` for some more details and flags.

### CLI

**Note**: this isn't really usable yet; most of the focus is on the web
interface for now.

    $ EXPORT PGINFO_DB='...'                # Set a default for the -db flag

    $ pg_info tables                        # Table overview.
    $ pg_info tables -table foo -detail     # List all columns for "foo".
    $ pg_info activity                      # pg_stat_activity.

    $ pg_info pev 'select 1'
    $ pg_info pev < query.sql

See `pg_info help` for some more commands and options for various commands.

### Add to Go app:

Right now it kind-of requires chi:

    r := chi.NewRouter()
    r.Mount("/admin/sql", pghandler.New(db))
