package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"zgo.at/pg_info"
	"zgo.at/pg_info/pghandler"
	"zgo.at/zdb"
	"zgo.at/zli"
)

/*
	h := zli.Helper()
	h.Add("", "adasd")
	h.Add("serve", "zxc")
*/

const help = `
Commands:

	help

	serve    

	activity
	indexes
	progress
	statements
	tables

	query
        -explain
        -pev
`

func main() {
	f := zli.NewFlags(os.Args)

	var (
		dbc    = f.String(os.Getenv("PGINFO_DB"), "db")
		listen = f.String(":9999", "listen")
	)

	err := f.Parse()
	zli.F(err)
	db, err := sql.Open("postgres", dbc.String())
	zli.F(err)
	zli.F(db.Ping())

	cmd := f.Shift()
	switch cmd {
	default:
		zli.Fatalf("unknown command: %q", cmd)
	case "", "help":
		fmt.Print(help)
	case "serve":
		s := http.Server{Addr: listen.String(), Handler: pghandler.New("", db)}
		fmt.Println("serving on", listen.String())
		zli.F(s.ListenAndServe())

	case "tables":
		ctx := zdb.With(context.Background(), sqlx.NewDb(db, "postgres"))

		var tbl pg_info.Tables
		err := tbl.Data(ctx)
		zli.F(err)

		for _, t := range tbl {
			fmt.Println(structPrint(t))
		}
	case "activity":
	}
}

func structPrint(s interface{}) string {
	typ := reflect.TypeOf(s)
	val := reflect.ValueOf(s)

	longest := 0
	for i := 0; i < typ.NumField(); i++ {
		t := typ.Field(i).Name
		if len(t) > longest {
			longest = len(t)
		}
	}

	b := new(strings.Builder)
	for i := 0; i < typ.NumField(); i++ {
		t := typ.Field(i).Name
		fmt.Fprintf(b, "%s: %s%v\n", t, strings.Repeat(" ", longest-len(t)), val.Field(i))
	}
	return b.String()
}
