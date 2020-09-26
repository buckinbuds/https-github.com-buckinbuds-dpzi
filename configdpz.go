// Copyright © 2019 Harrison Brown harrybrown98@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
)

type configCmd struct {
	cli.CliCommand
	cli.AddressBuilder
	db   *cache.DataBase
	conf *cli.Config

	file bool
	dir  bool
	edit bool
}

func (c *configCmd) Run(cmd *cobra.Command, args []string) error {
	if c.edit {
		return config.Edit()
	}
	if c.file {
		c.Println(config.File())
		return nil
	}
	if c.dir {
		c.Println(config.Folder())
		return nil
	}
	return config.FprintAll(cmd.OutOrStdout(), config.Object())
}

// NewConfigCmd creates a new config command.
func NewConfigCmd(b cli.Builder) cli.CliCommand {
	c := &configCmd{
		AddressBuilder: b,
		db:             b.DB(),
		conf:           b.Config(),
		file:           false,
		dir:            false,
	}
	c.CliCommand = b.Build("config", "Configure apizza", c)
	c.SetOutput(b.Output())
	cmd := c.Cmd()
	cmd.Aliases = []string{"conf"}
	cmd.Long = `The 'config' command is used for accessing the apizza config file
in your home directory. Feel free to edit the apizza config.json file
by hand or use the 'config' command.`

	c.Flags().BoolVarP(&c.file, "file", "f", c.file, "show the path to the config.json file")
	c.Flags().BoolVarP(&c.dir, "dir", "d", c.dir, "show the apizza config directory path")
	c.Flags().BoolVarP(&c.edit, "edit", "e", false, "open the config file with the text editor set by $EDITOR")

	cmd.AddCommand(configSetCmd, configGetCmd)
	return c
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "change variables in the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return set(args)
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "print the specified config variable to screen",
	RunE: func(cmd *cobra.Command, args []string) error {
		return get(args, cmd.OutOrStdout())
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func set(args []string) error {
	if len(args) < 1 {
		return errors.New("no variable given")
	}

	for _, arg := range args {
		keys := strings.Split(arg, "=")
		if len(keys) < 2 || keys[0] == "" || keys[1] == "" {
			return errors.New(`use '<key>=<value>' format (no spaces), use <key>='-' to set as empty`)
		}

		if keys[1] == "-" {
			keys[1] = ""
		}
		err := config.Set(keys[0], keys[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func get(args []string, out io.Writer) error {
	if len(args) < 1 {
		return errors.New("no variable given")
	}
	for _, arg := range args {
		v := config.Get(arg)
		if v == nil {
			return fmt.Errorf("cannot find %s", arg)
		}
		fmt.Fprintln(out, v)
	}
	return nil
}
