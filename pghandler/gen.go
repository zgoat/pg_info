// +build go_run_only

package main

import (
	"os"

	"zgo.at/zpack"
)

func main() {
	os.Chdir("..")
	err := zpack.Pack(map[string]map[string]string{
		"./pghandler/pack.go": {
			"packed": "tpl",
			"public": "public",
		},
	})
	if err != nil {
		panic(err)
	}
}
