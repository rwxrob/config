package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config "github.com/rwxrob/config/pkg"
	"github.com/rwxrob/fs/dir"
)

func ExampleDir() {

	dir := config.Dir("foo")
	parts := strings.Split(dir, string(filepath.Separator))
	fmt.Println(parts[len(parts)-2:])

	config.DefaultDir = `testdata`

	dir = config.Dir("foo")
	parts = strings.Split(dir, string(filepath.Separator))
	fmt.Println(parts[len(parts)-2:])

	// Output:
	// [.config foo]
	// [testdata foo]
}

func ExampleWrite() {
	config.DefaultDir = `testdata`

	thing := struct {
		Some  string
		Other string
	}{"some", "other"}
	if err := config.Write("foo", thing); err != nil {
		fmt.Println(err)
	}

	dir.Create(`testdata/foo`)
	defer os.RemoveAll(`testdata/foo`)

	if err := config.Write("foo", thing); err != nil {
		fmt.Println(err)
	}
	config.Print("foo")

	// Output:
	// some: some
	// other: other
}

func ExampleQuery() {
	config.DefaultDir = `testdata`

	config.QueryPrint("bar", ".")
	config.QueryPrint("bar", ".some")
	config.QueryPrint("bar", ".here")

	// Output:
	// {
	//   "command": {
	//     "path": "/here/we/go"
	//   },
	//   "here": "goes",
	//   "some": "thing"
	// }
	// thing
	// goes

}
