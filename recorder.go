package cmdtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/opts"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/harrybrwn/apizza/pkg/tests"
)

// Recorder is a mock command builder.
type Recorder struct {
	DataBase   *cache.DataBase
	Conf       *cli.Config
	Out        *bytes.Buffer
	cfgHasFile bool
	addr       dawg.Address
}

var services = []string{dawg.Carryout, dawg.Delivery}

// NewRecorder create a new command recorder.
func NewRecorder() *Recorder {
	addr := TestAddress()
	conf := &cli.Config{}
	config.DefaultOutput = ioutil.Discard
	err := config.SetConfig(".config/apizza/.tests", conf)
	if err != nil {
		panic(err.Error())
	}

	conf.Name = "Apizza TestRecorder"
	conf.Service = dawg.Carryout
	conf.Address = *addr

	out := new(bytes.Buffer)
	log.SetOutput(ioutil.Discard)

	return &Recorder{
		DataBase:   TempDB(),
		Out:        out,
		Conf:       conf,
		addr:       nil,
		cfgHasFile: true,
	}
}

// DB will return the internal database.
func (r *Recorder) DB() *cache.DataBase {
	return r.DataBase
}

// Config will return the config struct.
func (r *Recorder) Config() *cli.Config {
	return r.Conf
}

// Output returns the reqorder's output.
func (r *Recorder) Output() io.Writer {
	return r.Out
}

// Build a command.
func (r *Recorder) Build(use, short string, run cli.Runner) *cli.Command {
	c := cli.NewCommand(use, short, run.Run)
	c.SetOutput(r.Output())
	return c
}

// Address returns the address.
func (r *Recorder) Address() dawg.Address {
	if r.addr != nil {
		return r.addr
	}
	return &r.Conf.Address
}

// GlobalOptions has the global flags
func (r *Recorder) GlobalOptions() *opts.CliFlags {
	return &opts.CliFlags{}
}

// ToApp returns the arguments needed to create a cmd.App.
func (r *Recorder) ToApp() (*cache.DataBase, *cli.Config, io.Writer) {
	return r.DB(), r.Conf, r.Output()
}

// CleanUp will cleanup all the the Recorder tempfiles and free all resources.
func (r *Recorder) CleanUp() {
	var err error
	if r.cfgHasFile && config.File() != "" && config.Folder() != "" {
		err = config.Save()
		if err = config.Save(); err != nil {
			// panic(err)
			fmt.Println("Error:", err)
		}
		if err = os.Remove(config.File()); err != nil {
			// panic(err)
			fmt.Println("Error:", err)
		}
	}
	if err = r.DataBase.Destroy(); err != nil {
		// panic(err)
		fmt.Println("Error:", err)
	}
}

var _ cli.Builder = (*Recorder)(nil)

func must(db *cache.DataBase, e error) *cache.DataBase {
	if e != nil {
		panic(e)
	}
	return db
}

// ConfigSetup will set the internal recorder config to be main struct used
// in the config package.
func (r *Recorder) ConfigSetup(b []byte) error {
	return errs.Pair(
		config.SetNonFileConfig(r.Conf),
		json.Unmarshal(b, r.Conf),
	)
}

// Clear will clear all data stored by the recorder. This includes reseting
// the output buffer, opening a fresh database, and resetting the config.
func (r *Recorder) Clear() (err error) {
	r.ClearBuf()
	return r.FreshDB()
}

// ClearBuf will reset the internal output buffer.
func (r *Recorder) ClearBuf() {
	r.Out.Reset()
}

// FreshDB will close the old database, delete it, and open a fresh one.
func (r *Recorder) FreshDB() error {
	var err2 error
	err1 := r.DataBase.Destroy()
	r.DataBase, err2 = cache.GetDB(tests.NamedTempFile("test", "apizza_test.db"))
	return errs.Pair(err1, err2)
}

// Contains will return true if s is contained within the output buffer
// of the Recorder.
func (r *Recorder) Contains(s string) bool {
	return strings.Contains(r.Out.String(), s)
}

// StrEq compares a string with the recorder output buffer.
func (r *Recorder) StrEq(s string) bool {
	return r.Out.String() == s
}

// Compare the recorder output with a string
func (r *Recorder) Compare(t *testing.T, expected string) {
	tests.CompareCallDepth(t, r.Out.String(), expected, 2)
}

// TestRecorder is a Recorder that has access to a testing.T
type TestRecorder struct {
	*Recorder
	t *testing.T
}

// NewTestRecorder creates a new TestRecorder
func NewTestRecorder(t *testing.T) *TestRecorder {
	tests.InitHelpers(t)
	tr := &TestRecorder{
		Recorder: NewRecorder(),
		t:        t,
	}
	tr.init()
	return tr
}

var _ cli.Builder = (*TestRecorder)(nil)
