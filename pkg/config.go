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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rwxrob/fs"
	"github.com/rwxrob/fs/dir"
	"github.com/rwxrob/fs/file"
	"github.com/rwxrob/fs/lockedfile"
	y2j "github.com/rwxrob/y2j/pkg"
	"gopkg.in/yaml.v3"
)

// DefaultFile is name of the file within the configuration directory
// (DefaultDir).
var DefaultFile = "config.yaml"

// Dir is set to the os.UserConfigDir by default but can be changed for
// testing, etc.
var DefaultDir string

func init() {
	DefaultDir, _ = os.UserConfigDir()
}

// fulfills bonzai/conf.Configurer
type Configurer struct{}

func (Configurer) Init(id string) error          { return Init(id) }
func (Configurer) Data(id string) string         { return Data(id) }
func (Configurer) Print(id string)               { Print(id) }
func (Configurer) Edit(id string) error          { return Edit(id) }
func (Configurer) Write(id string, it any) error { return Write(id, it) }
func (Configurer) Query(id, q string) string     { return Query(id, q) }
func (Configurer) QueryPrint(id, q string)       { QueryPrint(id, q) }

// Init initializes the configuration directory for the current user and
// given application name (id) using the standard os.UserConfigDir location.
// The directory is completely removed and new configuration file(s) are
// created. Consider placing a confirmation prompt before calling this
// function when term.IsInteractive is true. Since Init uses
// fs/{dir,file}.Create you can set the file.DefaultPerms and
// dir.DefaultPerms if you prefer a different default for your
// permissions. Permissions in the fs package are restrictive
// (0700/0600) by default to  allow tokens to be stored within
// configuration files (as other applications are known to do). Still,
// saving of critical secrets is not encouraged within any flat
// configuration file. But anything that a web browser would need to
// cache in order to operate is appropriate (cookies, session tokens,
// etc.).
func Init(id string) error {
	d := Dir(id)
	if d == "" {
		return fmt.Errorf("could not resolve config path for %q", id)
	}
	if len(id) == 0 && len(d) == 0 {
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
	return file.Touch(filepath.Join(d, DefaultFile))
}

// Dir simply joins os.UserConfigDir to the id returning No check for
// the existence of the directory is made. An empty string is returned
// if there are any errors.
func Dir(id string) string {
	return filepath.Join(DefaultDir, id)
}

// File returns the full path to the specified DefaultFile for the
// given configuration.
func File(id string) string {
	return filepath.Join(DefaultDir, id, DefaultFile)
}

// Data returns a string buffer containing all of the configuration file
// data for the given configuration. An empty string is returned if any
// error occurs.
func Data(id string) string {
	buf, _ := os.ReadFile(File(id))
	return string(buf)
}

// Print prints the Data to standard output with an additional line
// return.
func Print(id string) { fmt.Println(Data(id)) }

// Edit opens the given configuration the local editor. See fs/file.Edit
// for more.
func Edit(id string) error {
	if err := mkdir(id); err != nil {
		return err
	}
	path := File(id)
	if path == "" {
		return fmt.Errorf("unable to locate config for %q", id)
	}
	return file.Edit(path)
}

func mkdir(id string) error {
	d := Dir(id)
	if d == "" {
		return fmt.Errorf("failed to find config for %q", id)
	}
	if fs.NotExists(d) {
		if err := dir.Create(d); err != nil {
			return err
		}
	}
	return nil
}

// Write marshals any Go type and overwrites the configuration File in
// a way that is safe for all callers of Write in this current system for
// any operating system using fs/lockedfile (taken from the to internal
// project itself, https://github.com/golang/go/issues/33974) but
// applying the file.DefaultPerms instead of the 0666 Go default.
func Write(id string, conf any) error {
	buf, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	if err := mkdir(id); err != nil {
		return err
	}
	return lockedfile.Write(File(id),
		bytes.NewReader(buf), _fs.FileMode(file.DefaultPerms))
}

// Query returns a JSON string matching the given jq-like query for the
// given configuration. Currently, this function is implemented by
// calling jq on the host system, but eventually it will be ported to
// use a version of jq written in Go natively. Will log an error and
// return empty string if error.
func Query(id, q string) string {

	data := Data(id)
	if data == "" {
		return ""
	}

	// TODO: replace this with yq
	datab, err := y2j.Convert([]byte(data))
	if err != nil {
		//log.Print(err)
		return ""
	}

	buf := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command("jq", "-r", q)
	cmd.Stdin = bytes.NewReader(datab)
	cmd.Stdout = buf
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		//log.Print(stderr)
		//log.Print(err)
		return ""
	}

	return strings.TrimSpace(buf.String())
}

// QueryPrint prints the output of Query with a new line.
func QueryPrint(id, q string) { fmt.Println(Query(id, q)) }
