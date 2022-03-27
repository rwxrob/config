package config

import (
	"fmt"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/inc/help"
	config "github.com/rwxrob/config/pkg"
	"github.com/rwxrob/term"
)

var Cmd = &bonzai.Cmd{

	Name:      `config`,
	Summary:   `manage local YAML/JSON configuation`,
	Version:   `v0.0.1`,
	Copyright: `Copyright 2021 Robert S Muhlestein`,
	License:   `Apache-2.0`,
	Commands:  []*bonzai.Cmd{data, _init, edit, _file, query, help.Cmd},
	Description: `
		The config command allows configuration of the current command in
		YAML and JSON (since all JSON is valid YAML). All changes to
		configuration values are done via the <edit> command since
		configurations may be complex and deeply nested in some cases.
		Querying configuration data, however, can be easily accomplished
		with the <query> command that uses jq-like selection syntax.`,
}

var _init = &bonzai.Cmd{
	Name:     `init`,
	Aliases:  []string{"i"},
	Summary:  `(re)initializes the current configuration cache`,
	Commands: []*bonzai.Cmd{help.Cmd},
	Call: func(x *bonzai.Cmd, _ ...string) error {
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

var _file = &bonzai.Cmd{
	Name:     `file`,
	Aliases:  []string{"f"},
	Summary:  `outputs the full path to the configuration file`,
	Commands: []*bonzai.Cmd{help.Cmd},
	Call: func(x *bonzai.Cmd, _ ...string) error {
		path := config.File(x.Root.Name)
		if path == "" {
			return fmt.Errorf("unable to file config for %q",
				x.Root.Name)
		}
		fmt.Println(path)
		return nil
	},
}

var data = &bonzai.Cmd{
	Name:     `data`,
	Aliases:  []string{"d"},
	Summary:  `outputs the contents of the configuration file`,
	Commands: []*bonzai.Cmd{help.Cmd},
	Call: func(x *bonzai.Cmd, _ ...string) error {
		data := config.Data(x.Root.Name)
		if data == "" {
			return fmt.Errorf("config empty or missing for %q",
				x.Root.Name)
		}
		fmt.Print(data)
		return nil
	},
}

var edit = &bonzai.Cmd{
	Name:     `edit`,
	Summary:  `edit config in user home config location`,
	Aliases:  []string{"e"},
	Commands: []*bonzai.Cmd{help.Cmd},

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

	Call: func(x *bonzai.Cmd, _ ...string) error {
		return config.Edit(x.Root.Name)
	},
}

var query = &bonzai.Cmd{
	Name:     `query`,
	Summary:  `query configuration data using jq/yq style`,
	Usage:    `<dotted>`,
	Aliases:  []string{"q"},
	Commands: []*bonzai.Cmd{help.Cmd},
	Call: func(x *bonzai.Cmd, args ...string) error {
		if len(args) == 0 {
			return x.UsageError()
		}
		config.QueryPrint(x.Root.Name, args[0])
		return nil
	},
}
