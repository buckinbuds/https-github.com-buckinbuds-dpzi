package opts

import (
	"time"

	"github.com/spf13/pflag"
)

// MenuUpdateTime is the time a menu is persistant in cache
const MenuUpdateTime = 12 * time.Hour

// CliFlags for the root apizza command.
type CliFlags struct {
	Address string
	Service string// Copyright © 2019 Harrison Brown harrybrown98@gmail.com
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

package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	fp "path/filepath"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/commands"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger for the cmd package
var Logger = &lumberjack.Logger{
	Filename:   "",
	MaxSize:    25,  // megabytes
	MaxBackups: 10,  // number of spare files
	MaxAge:     365, //days
	Compress:   false,
}

var (
	// Version is the cli version id (will be set as an ldflag)
	version string

	// testing version change this with an ldflag
	enableLog = "yes"
)

// AllCommands returns a list of all the Commands.
func AllCommands(builder cli.Builder) []*cobra.Command {
	return []*cobra.Command{
		commands.NewCartCmd(builder).Cmd(),
		commands.NewConfigCmd(builder).Cmd(),
		NewMenuCmd(builder).Cmd(),
		commands.NewOrderCmd(builder).Cmd(),
		commands.NewAddAddressCmd(builder, os.Stdin).Cmd(),
		commands.NewCompletionCmd(builder),
	}
}

// Execute runs the root command
func Execute(args []string, dir string) (msg *ErrMsg) {
	app := NewApp(os.Stdout)
	err := app.Init(dir)
	if err != nil {
		return senderr(err, "Internal Error", 1)
	}

	if enableLog == "yes" {
		Logger.Filename = fp.Join(config.Folder(), "logs", "dev.log")
		log.SetOutput(Logger)
	}

	defer func() {
		errmsg := senderr(app.Cleanup(), "Internal Error", 1)
		if errmsg != nil {
			// if we always set it the the return value will always
			// be the same as errmsg
			msg = errmsg
		}
	}()

	cmd := app.Cmd()
	cmd.Version = version
	cmd.SetArgs(args)
	cmd.AddCommand(AllCommands(app)...)
	return senderr(cmd.Execute(), "Error", 1)
}

// ErrMsg is not actually an error but it is my way of
// containing an error with a message and an exit code.
type ErrMsg struct {
	Msg  string
	Code int
	Err  error
}

func senderr(e error, msg string, code int) *ErrMsg {
	if e == nil {
		return nil
	}
	return &ErrMsg{Msg: msg, Code: code, Err: e}
}

var test = false

func newTestCmd(b cli.Builder, valid bool) *cobra.Command {
	return &cobra.Command{
		Use:    "test",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !valid {
				return errors.New("no such command 'test'")
			}

			db := b.DB()
			fmt.Printf("%+v\n", db)

			m, _ := db.Map()
			for k := range m {
				fmt.Println(k)
			}
			return nil
		},
	}
}


	ClearCache bool
	ResetMenu  bool
	LogFile    string
}

// Install the RootFlags
func (rf *CliFlags) Install(persistflags *pflag.FlagSet) {
	rf.ClearCache = false
	persistflags.BoolVar(&rf.ResetMenu, "delete-menu", false, "delete the menu stored in cache")
	persistflags.StringVar(&rf.LogFile, "log", "", "set a log file (found in ~/.config/apizza/logs)")

	persistflags.StringVarP(&rf.Address, "address", "A", rf.Address, "an address name stored with 'apizza address --new'")
	persistflags.StringVar(&rf.Service, "service", rf.Service, "select a Dominos service, either 'Delivery' or 'Carryout'")
}

// ApizzaFlags that are not persistant.
type ApizzaFlags struct {
	StoreLocation bool

	// developer opts
	Openlogs bool
	Dumpdb   bool
}

// Install the apizza flags
func (af *ApizzaFlags) Install(flags *pflag.FlagSet) {
	flags.BoolVarP(&af.StoreLocation, "store-location", "L", false, "show the location of the nearest store")

	flags.BoolVar(&af.Openlogs, "open-logs", false, "open the log file")
	flags.MarkHidden("open-logs")
	flags.BoolVar(&af.Dumpdb, "dump-db", false, "dump the database to stdout as json")
	flags.MarkHidden("dump-db")
}
