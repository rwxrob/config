# 🌳 Go YAML/JSON Configuration

⛔ **DEPRECATION WARNING:** Use [conf](https://github.com/rwxrob/conf) instead.

![Go
Version](https://img.shields.io/github/go-mod/go-version/rwxrob/config)
[![GoDoc](https://godoc.org/github.com/rwxrob/config?status.svg)](https://godoc.org/github.com/rwxrob/config)
[![License](https://img.shields.io/badge/license-Apache2-brightgreen.svg)](LICENSE)

This `config` Bonzai branch is for safely managing any configuration as
single, local YAML/JSON using industry standards for local configuration
and system-safe writes. Use it to add a `config` subcommand to any other
Bonzai command, or to your root Bonzai tree (`z`). All commands that use
`config` that are composed into a single binary, no matter where in the
tree, will use the same local config file even though the position
within the file will be qualified by the tree location.

By default, importing `config` will assigned a new implementation of
`bonzai.Configurer` to `Z.Conf` (satisfying any `Z.Cmd` requirement for
configuration) and will use the name of the binary (`Z.ExeName`) as the
directory name within `os.UserConfDir` with a `config.yaml` file name.
To override this behavior, create a new `pkd/config.Conf` struct assign
`Id`, `Dir` and `File`, and then assign that to `Z.Conf`.

## Install

This command can be installed as a standalone program (for combined use
with shell scripts perhaps) or composed into a Bonzai command tree.

Standalone

```
go install github.com/rwxrob/config/config@latest
```

Composed

```go
package z

import (
	Z "github.com/rwxrob/bonzai"
	"github.com/rwxrob/config"
)

var Cmd = &bonzai.Cmd{
	Name:     `z`,
	Commands: []*Z.Cmd{help.Cmd, config.Cmd},
}
```

Note config is designed to be composed only in monolith mode (not
multicall binary).

## Tab Completion

To activate bash completion just use the `complete -C` option from your
`.bashrc` or command line. There is no messy sourcing required. All the
completion is done by the program itself.

```
complete -C config config
```

If you don't have bash or tab completion check use the shortcut
commands instead.

## Embedded Documentation

All documentation (like manual pages) has been embedded into the source
code of the application. See the source or run the program with help to
access it.

## Design Considerations

* **JSON Output.** JSON is YAML. But JSON is also much safer to deal
  with when parsing and piping into other things. The `Query` form has
  been modeled after `jq` (which has become something of a standard tool
  for mining information from configuration and other files.
