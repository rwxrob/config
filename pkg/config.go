/*
Package config helps generically manage configuration data in YAML files
(and, by extension JSON, which is a YAML subset) using the
gopkg.in/yaml.v3 package (v2 is not compatible with encoding/json
creating unexpected marshaling errors).

The package provides the high-level functions called by the Bonzai™
branch command of the same name allowing it to be consistently composed into any to any other Bonzai branch. However, composing into the root is generally preferred to avoid configuration name space conflicts and centralize all configuration data for a single Bonzai tree monolith for easy transport.

Rather than provide functions for changing individual values, this
package assumes editing of the YAML files directly and provider helpers
for system-wide safe concurrent writes and queries using the convention
yq/jq syntax. Default directory and file permissions are purposefully
more restrictive than the Go default (0700/0600).
*/
package config

import (
	"bytes"
	"fmt"
	_fs "io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/rogpeppe/go-internal/lockedfile"
	"github.com/rwxrob/fs"
	"github.com/rwxrob/fs/dir"
	"github.com/rwxrob/fs/file"
	yq "github.com/rwxrob/yq/pkg"
	"gopkg.in/yaml.v3"
)

// Conf fulfills bonzai.Configurer
type Conf struct {
	Id   string // usually application name
	Dir  string // usually os.UserConfigDir
	File string // usually config.yaml
}

// DirPath is the Dir and Id joined.
func (c Conf) DirPath() string { return filepath.Join(c.Dir, c.Id) }

// Path returns the combined Dir and File.
func (c Conf) Path() string { return filepath.Join(c.Dir, c.Id, c.File) }

// Init initializes the configuration directory (Dir) for the current
// user and given application name (Id) using the standard
// os.UserConfigDir location.  The directory is completely removed and
// new configuration file(s) are created.
//
// Consider placing a confirmation prompt before calling this function
// when term.IsInteractive is true. Since Init uses fs/{dir,file}.Create
// you can set the file.DefaultPerms and dir.DefaultPerms if you prefer
// a different default for your permissions.
//
// Permissions in the fs package are restrictive (0700/0600) by default
// to  allow tokens to be stored within configuration files (as other
// applications are known to do). Still, saving of critical secrets is
// not encouraged within any flat configuration file. But anything that
// a web browser would need to cache in order to operate is appropriate
// (cookies, session tokens, etc.).
func (c Conf) Init() error {
	d := c.DirPath()

	// safety checks before blowing things away
	if d == "" {
		return fmt.Errorf("could not resolve config path for %q", c.Id)
	}
	if len(c.Id) == 0 && len(c.Dir) == 0 {
		return fmt.Errorf("empty directory id")
	}

	if fs.Exists(d) {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}

	if err := dir.Create(d); err != nil {
		return err
	}

	return file.Touch(c.Path())
}

// Data returns a string buffer containing all of the configuration file
// data for the given configuration. An empty string is returned and an
// error logged if any error occurs.
func (c Conf) Data() string {
	buf, err := os.ReadFile(c.Path())
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(buf)
}

// Print prints the Data to standard output with an additional line
// return.
func (c Conf) Print() { fmt.Println(c.Data()) }

// Edit opens the given configuration the local editor. See fs/file.Edit
// for more.
func (c Conf) Edit() error {
	if err := c.mkdir(); err != nil {
		return err
	}
	path := c.Path()
	if path == "" {
		return fmt.Errorf("unable to locate config for %q", c.Id)
	}
	return file.Edit(path)
}

func (c Conf) mkdir() error {
	d := c.DirPath()
	if d == "" {
		return fmt.Errorf("failed to find config for %q", c.Id)
	}
	if fs.NotExists(d) {
		if err := dir.Create(d); err != nil {
			return err
		}
	}
	return nil
}

// OverWrite marshals any Go type and overwrites the configuration File in
// a way that is safe for all callers of OverWrite in this current system
// for any operating system using go-internal/lockedfile (taken from the
// to internal project itself,
// https://github.com/golang/go/issues/33974) but applying the
// file.DefaultPerms instead of the 0666 Go default.
func (c Conf) OverWrite(newconf any) error {
	buf, err := yaml.Marshal(newconf)
	if err != nil {
		return err
	}
	if err := c.mkdir(); err != nil {
		return err
	}
	return lockedfile.Write(c.Path(),
		bytes.NewReader(buf), _fs.FileMode(file.DefaultPerms))
}

// Query returns a YAML string matching the given yq query for the given
// configuration. Currently, this function is implemented by calling
// rwxrob/yq. Will log and return empty string if error.
func (c Conf) Query(q string) string {
	results, err := yq.EvaluateToString(q, c.Path())
	if err != nil {
		log.Print(err)
		return ""
	}
	return results
}

// QueryPrint prints the output of Query with a new line.
func (c Conf) QueryPrint(q string) { fmt.Print(c.Query(q)) }
