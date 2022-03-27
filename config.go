/*
Package config helps generically manage configuration data in YAML files
(and, by extension JSON, which is a YAML subset) using the
gopkg.in/yaml.v3 package (v2 is not compatible with encoding/json
creating unexpected marshaling errors).

The package provides the high-level functions called by the Bonzai™
branch command of the same name allowing it to be consistently composed into any to any other Bonzai branch.

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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rwxrob/fs"
	"github.com/rwxrob/fs/dir"
	"github.com/rwxrob/fs/file"
	"github.com/rwxrob/fs/lockedfile"
	"github.com/rwxrob/yaml2json"
	"gopkg.in/yaml.v2"
)

// DefaultFile is name of the file within the configuration directory
// (DefaultDir).
var DefaultFile = "config.yaml"

// Dir is set to the os.UserConfigDir by default but can be changed for
// testing, etc.
var DefaultDir string

func init() {
	var err error
	DefaultDir, err = os.UserConfigDir()
	if err != nil {
		log.Print(err)
	}
}

// Init initializes the configuration directory for the current user and
// given application name using the standard os.UserConfigDir location.
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
func Init(name string) error {
	d := Dir(name)
	if d == "" {
		return fmt.Errorf("could not resolve config path for %q", name)
	}
	if len(name) == 0 && len(d) == 0 {
		return fmt.Errorf("empty directory name")
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

// Dir simply joins os.UserConfigDir to the name returning No check for
// the existence of the directory is made. An empty string is returned
// if there are any errors.
func Dir(name string) string {
	return filepath.Join(DefaultDir, name)
}

// File returns the full path to the specified DefaultFile for the
// configuration named.
func File(name string) string {
	return filepath.Join(DefaultDir, name, DefaultFile)
}

// Data returns a string buffer containing all of the configuration file
// data for the named configuration. An empty string is returned if any
// error occurs.
func Data(name string) string {
	buf, _ := os.ReadFile(File(name))
	return string(buf)
}

// Print prints the Data to standard output with an additional line
// return.
func Print(name string) { fmt.Println(Data(name)) }

// Edit opens the named configuration the local editor. See fs/file.Edit
// for more.
func Edit(name string) error {
	if err := mkdir(name); err != nil {
		return err
	}
	path := File(name)
	if path == "" {
		return fmt.Errorf("unable to locate config for %q", name)
	}
	return file.Edit(path)
}

func mkdir(name string) error {
	d := Dir(name)
	if d == "" {
		return fmt.Errorf("failed to find config for %q", name)
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
func Write(name string, conf any) error {
	buf, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	if err := mkdir(name); err != nil {
		return err
	}
	return lockedfile.Write(File(name),
		bytes.NewReader(buf), _fs.FileMode(file.DefaultPerms))
}

// Query returns a JSON string matching the given jq-like query for the
// named configuration. Currently, this function is implemented by
// calling jq on the host system, but eventually it will be ported to
// use a version of jq written in Go natively. Will log an error and
// return empty string if error.
func Query(name, q string) string {

	data := Data(name)
	if data == "" {
		return ""
	}

	// TODO: replace this with yq
	datab, err := yaml2json.Convert([]byte(data))
	if err != nil {
		log.Print(err)
		return ""
	}

	buf := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command("jq", "-r", q)
	cmd.Stdin = bytes.NewReader(datab)
	cmd.Stdout = buf
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		log.Print(stderr)
		log.Print(err)
		return ""
	}

	return strings.TrimSpace(buf.String())
}

// QueryPrint prints the output of Query with a new line.
func QueryPrint(name, q string) { fmt.Println(Query(name, q)) }
