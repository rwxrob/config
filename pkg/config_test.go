package config_test

import (
	"fmt"
	"os"

	config "github.com/rwxrob/config/pkg"
	"github.com/rwxrob/fs/dir"
)

func ExampleConf_OverWrite() {

	c := config.Conf{Id: `foo`, Dir: `testdata`, File: `config.yaml`}

	thing := struct {
		Some  string
		Other string
	}{"some", "other"}

	if err := c.OverWrite(thing); err != nil {
		fmt.Println(err)
	}

	dir.Create(`testdata/foo`)
	defer os.RemoveAll(`testdata/foo`)

	if err := c.OverWrite(thing); err != nil {
		fmt.Println(err)
	}
	c.Print()

	// Output:
	// some: some
	// other: other
}

func ExampleQuery() {

	c := config.Conf{Id: `bar`, Dir: `testdata`, File: `config.yaml`}

	c.QueryPrint(".")
	fmt.Println()
	c.QueryPrint(".some")
	fmt.Println()
	c.QueryPrint(".here")

	// Output:
	// some: thing
	// here: goes
	// command:
	//   path: /here/we/go
	// thing
	// goes

}
