package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/harrybrwn/apizza/pkg/tests"
)

var testconfigjson = `
{
	"name":"joe","email":"nojoe@mail.com", "phone":"1231231234",
	"address":{
		"street":"1600 Pennsylvania Ave NW",
		"cityName":"Washington DC",
		"state":"","zipcode":"20500"
	},
	"default-address-name": "",
	"card":{"number":"","expiration":"","cvv":""},
	"service":"Carryout"
}`

var testConfigOutput = `name: "joe"
email: "nojoe@mail.com"
phone: "1231231234"
address:
  street: "1600 Pennsylvania Ave NW"
  cityname: "Washington DC"
  state: ""
  zipcode: "20500"
default-address-name: ""
card:
  number: ""
  expiration: ""
service: "Carryout"
`

func TestConfigStruct(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	r.ConfigSetup([]byte(testconfigjson))
	defer func() { r.CleanUp() }()
	tests.Fatal(json.Unmarshal([]byte(testconfigjson), r.Config()))

	tests.StrEq(r.Config().Get("name").(string), "joe", "wrong value from Config.Get")
	tests.Check(r.Config().Set("name", "not joe"))
	tests.StrEq(r.Config().Get("Name").(string), "not joe", "wrong value from Config.Get")
	tests.Check(r.Config().Set("name", "joe"))
}

func TestConfigCmd(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	c := NewConfigCmd(r).(*configCmd)
	r.ConfigSetup([]byte(testconfigjson))
	defer r.CleanUp()

	c.file = true
	tests.Check(c.Run(c.Cmd(), []string{}))
	c.file = false
	r.Compare(t, "\n")
	r.ClearBuf()
	c.dir = true
	tests.Check(c.Run(c.Cmd(), []string{}))
	r.Compare(t, "\n")
	r.ClearBuf()

	tests.Check(json.Unmarshal([]byte(testconfigjson), r.Config()))
	c.dir = false
	tests.Check(c.Run(c.Cmd(), []string{}))
	r.Compare(t, testConfigOutput)
	r.ClearBuf()
}

func TestConfigEdit(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	c := NewConfigCmd(r).(*configCmd)
	tests.Check(config.SetConfig(".config/apizza/tests", r.Conf))
	defer func() {
		tests.Check(errs.Pair(r.DB().Destroy(), os.RemoveAll(config.Folder())))
	}()

	os.Setenv("EDITOR", "cat")
	c.edit = true
	exp := `{
    "Name": "",
    "Email": "",
    "Phone": "",
    "Address": {
        "Street": "",
        "CityName": "",
        "State": "",
        "Zipcode": ""
    },
    "DefaultAddressName": "",
    "Card": {
        "Number": "",
        "Expiration": ""
    },
    "Service": "Delivery"
}`
	if exp == "" {
		t.Error("no this should not happed")
	}
	t.Run("edit output", func(t *testing.T) {
		if os.Getenv("TRAVIS") != "true" {
			// for some reason, 'cat' in travis gives no output
			// tests.CompareOutput(t, exp, func() {
			// tests.Check(c.Run(c.Cmd(), []string{}))
			// fmt.Println(exp)
			// })
		}
	})
	c.edit = false

	tests.Check(json.Unmarshal([]byte(testconfigjson), r.Config()))
	a := config.Get("address")
	if a == nil {
		t.Error("should not be nil")
	}
	addr, ok := a.(obj.Address)
	if !ok {
		t.Error("this should be an obj.Address")
	}
	if addr.City() != "Washington DC" {
		t.Error("bad config address city")
	}
}

func TestConfigGet(t *testing.T) {
	tests.InitHelpers(t)
	conf := &cli.Config{}
	config.SetNonFileConfig(conf) // don't want it to over ride the file on disk
	tests.Fatal(json.Unmarshal([]byte(testconfigjson), conf))
	buf := &bytes.Buffer{}

	args := []string{"email", "name"}
	tests.Check(get(args, buf))
	tests.Check(configGetCmd.RunE(configGetCmd, []string{"email", "name"}))
	tests.Compare(t, buf.String(), "nojoe@mail.com\njoe\n")
	buf.Reset()

	args = []string{}
	err := get(args, buf)
	if err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}

	args = []string{"nonExistantKey"}
	err = get(args, buf)
	if err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}

	if err := configGetCmd.RunE(configGetCmd, []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}
	if err := configGetCmd.RunE(configGetCmd, []string{"nonExistantKey"}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func TestConfigSet(t *testing.T) {
	tests.InitHelpers(t)
	conf := &cli.Config{}
	config.SetNonFileConfig(conf) // don't want it to over ride the file on disk
	tests.Fatal(json.Unmarshal([]byte(cmdtest.TestConfigjson), conf))

	tests.Check(configSetCmd.RunE(configSetCmd, []string{"name=someNameOtherThanJoe"}))
	tests.StrEq(config.GetString("name"), "someNameOtherThanJoe", "did not set the name correctly")
	if err := configSetCmd.RunE(configSetCmd, []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}
	if err := configSetCmd.RunE(configSetCmd, []string{"nonExistantKey=someValue"}); err == nil {
		t.Error("expected error")
	}
	if err := configSetCmd.RunE(configSetCmd, []string{"badformat"}); err == nil {
		t.Error(err)
	} else if err.Error() != "use '<key>=<value>' format (no spaces), use <key>='-' to set as empty" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func TestCompletion(t *testing.T) {
	r := cmdtest.NewTestRecorder(t)
	defer r.CleanUp()
	c := NewCompletionCmd(r)
	c.SetOutput(r.Out)
	for _, a := range []string{"zsh", "ps", "powershell", "bash", "fish"} {
		if err := c.RunE(c, []string{a}); err != nil {
			if r.Out.Len() == 0 {
				t.Error("got zero length completion script")
			}
		}
	}
}

func TestAddressCmd(t *testing.T) {
	r := cmdtest.NewTestRecorder(t)
	defer r.CleanUp()
	buf := &bytes.Buffer{}
	cmd := NewAddAddressCmd(r, buf).(*addAddressCmd)
	tests.Check(cmd.Cmd().ParseFlags([]string{"--new"}))

	inputs := []string{
		"testaddress",
		"600 Mountain Ave bldg 5",
		"New Providence",
		"NJ",
		"07974",
	}
	for _, in := range inputs {
		_, err := buf.Write([]byte(in + "\n"))
		tests.Check(err)
	}
	tests.Check(cmd.Run(cmd.Cmd(), []string{}))
	raw, err := r.DataBase.WithBucket("addresses").Get("testaddress")
	tests.Check(err)
	addr, err := obj.FromGob(raw)
	tests.Check(err)
	tests.StrEq(addr.Street, "600 Mountain Ave bldg 5", "got wrong street")
	tests.StrEq(addr.CityName, "New Providence", "got wrong city")
	tests.StrEq(addr.State, "NJ", "go wrong state")
	tests.StrEq(addr.Zipcode, "07974", "got wrong zip")

	r.Out.Reset()
	cmd.new = false
	tests.Check(cmd.Run(cmd.Cmd(), []string{}))
	if !strings.Contains(r.Out.String(), obj.AddressFmtIndent(addr, 2)) {
		t.Error("address was no found in output")
	}
	r.Out.Reset()

	tests.Check(cmd.Cmd().ParseFlags([]string{"--delete=testaddress"}))
	tests.Check(cmd.Run(cmd.Cmd(), []string{}))
	if r.Out.Len() != 0 {
		t.Error("should be zero length")
	}
}
