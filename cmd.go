package config

import (
	"fmt"

	"github.com/rwxrob/bonzai/help"
	Z "github.com/rwxrob/bonzai/z"
	config "github.com/rwxrob/config/pkg"
	"github.com/rwxrob/term"
)

var Cmd = &Z.Cmd{

	Name:      `config`,
	Summary:   `manage local YAML/JSON configuation`,
	Version:   `v0.0.1`,
	Copyright: `Copyright 2021 Robert S Muhlestein`,
	License:   `Apache-2.0`,
	Commands:  []*Z.Cmd{data, _init, edit, _file, query, help.Cmd},
	Description: `
		The *config* Bonzai branch is for safely managing any configuration
		as single, local YAML/JSON using industry standards for local
		configuration. Use it to add a *config* subcommand to any other
		Bonzai command, or to your root Bonzai tree (*z*).

		Take particular note that all commands composed into a single
		binary, no matter where in the tree, will use the same local config
		file even though the position within the file will be qualified by
		the tree location. Therefore, any composite command can read the
		configurations of any other composite command within the same
		binary. This is by design, but all commands composed together should
		always be vetted for safe practices. This is also the reason there
		is no "write" or "set" command.

		All changes to configuration values are done via the *edit* command
		since configurations may be complex and deeply nested in some cases
		and promoting the automatic changing of configuration values opens
		the possibility of one buggy composed command to blow away one or
		all the configurations for everything composed into the binary. [The
		*cache* command is recommended when wanting to persist local data
		between command executions.]

		Querying configuration data can be easily accomplished with the
		<query> command that uses jq-like selection syntax.`,
}

var _init = &Z.Cmd{
	Name:     `init`,
	Aliases:  []string{"i"},
	Summary:  `(re)initializes the current configuration cache`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		if term.IsInteractive() {
			dir := config.Dir(x.Root.Name)
			if dir == "" {
				return fmt.Errorf("unable to resolve config for %q", x.Root.Name)
			}
			r := term.Prompt(`Really initialize %v? (y/N) `, dir)
			if r != "y" {
				return nil
			}
		}
		return config.Init(x.Root.Name)
	},
}

var _file = &Z.Cmd{
	Name:     `file`,
	Aliases:  []string{"f"},
	Summary:  `outputs the full path to the configuration file`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		path := config.File(x.Root.Name)
		if path == "" {
			return fmt.Errorf("unable to file config for %q",
				x.Root.Name)
		}
		fmt.Println(path)
		return nil
	},
}

var data = &Z.Cmd{
	Name:     `data`,
	Aliases:  []string{"d"},
	Summary:  `outputs the contents of the configuration file`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		data := config.Data(x.Root.Name)
		if data == "" {
			return fmt.Errorf("config empty or missing for %q",
				x.Root.Name)
		}
		fmt.Print(data)
		return nil
	},
}

var edit = &Z.Cmd{
	Name:     `edit`,
	Summary:  `edit config in user home config location`,
	Aliases:  []string{"e"},
	Commands: []*Z.Cmd{help.Cmd},

	Description: `
		The edit command will the configuration file for editing in the
		currently configured editor (in order or priority):

		* $VISUAL
		* $EDITOR
		* vi
		* vim
		* nano

		The edit command hands over control of the currently running process
		to the editor. `,

	Call: func(x *Z.Cmd, _ ...string) error {
		return config.Edit(x.Root.Name)
	},
}

var query = &Z.Cmd{
	Name:     `query`,
	Summary:  `query configuration data using jq/yq style`,
	Usage:    `<dotted>`,
	Aliases:  []string{"q"},
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, args ...string) error {
		if len(args) == 0 {
			return x.UsageError()
		}
		config.QueryPrint(x.Root.Name, args[0])
		return nil
	},
}
