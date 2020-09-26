package dawg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func testServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	m := http.NewServeMux()
	srv := httptest.NewServer(m)
	u, err := url.Parse(srv.URL)
	tr := &TestTransport{
		host: u.Host,
		rt: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return u, err
			},
		}}
	c := &http.Client{Transport: tr}
	return c, m, srv
}

type TestTransport struct {
	host string
	rt   http.RoundTripper
}

func (tt *TestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = "http"
	if r.URL.Host == "" {
		r.URL.Host = tt.host
	}
	return tt.rt.RoundTrip(r)
}

func swapClientWith(c *http.Client) func() {
	old := orderClient
	orderClient = &client{Client: c}
	return func() { orderClient = old }
}

func TestFormat(t *testing.T) {
	url := format("https://order.dominos.com/power/%s", "store-locator")
	expected := "https://order.dominos.com/power/store-locator"
	if url != expected {
		t.Errorf("Expected: %s, Got: %s", expected, url)
	}
}

func TestOrderAddressConvertion(t *testing.T) {
	tests.InitHelpers(t)
	exp := &StreetAddr{StreetNum: "1600", StreetName: "Pennsylvania Ave.",
		Street: "1600 Pennsylvania Ave.", CityName: "Washington",
		State: "DC", Zipcode: "20500", AddrType: "House"}

	addr := &UserAddress{
		Street:     "1600 Pennsylvania Ave.",
		CityName:   "Washington",
		PostalCode: "20500",
		Region:     "DC",
	}

	res := StreetAddrFromAddress(addr)
	tests.StrEq(res.City(), exp.City(), "wrong city")
	tests.StrEq(res.LineOne(), exp.LineOne(), "wrong lineone")
	tests.StrEq(res.StateCode(), exp.StateCode(), "wrong state code")
	tests.StrEq(res.Zip(), exp.Zip(), "wrong zip code")
	tests.StrEq(res.StreetNum, exp.StreetNum, "wrong street number")
	tests.StrEq(res.StreetName, exp.StreetName, "wrong street name")
}

func TestParseAddressTable(t *testing.T) {
	tests.InitHelpers(t)
	var cases = []struct {
		raw      string
		expected StreetAddr
	}{
		{
			raw: `1600 Pennsylvania Ave. Washington, DC 20500`,
			expected: StreetAddr{StreetNum: "1600", StreetName: "Pennsylvania Ave.",
				Street: "1600 Pennsylvania Ave.", CityName: "Washington",
				State: "DC", Zipcode: "20500", AddrType: "House"},
		},
		{
			raw: `378 James St. Chicago, IL 60621`,
			expected: StreetAddr{StreetNum: "378", StreetName: "James St.",
				Street: "378 James St.", CityName: "Chicago", State: "IL",
				Zipcode: "60621"},
		},
	}

	for _, tc := range cases {
		addr, err := ParseAddress(tc.raw)
		tests.Check(err)
		tests.StrEq(addr.StreetNum, tc.expected.StreetNum, "wrong street num")
		tests.StrEq(addr.Street, tc.expected.Street, "wrong street")
		tests.StrEq(addr.CityName, tc.expected.CityName, "wrong city")
		tests.StrEq(addr.State, tc.expected.State, "wrong state")
		tests.StrEq(addr.Zipcode, tc.expected.Zipcode, "wrong zip")
	}
}

func TestNetworking_Err(t *testing.T) {
	t.Skip("this test takes way too long")
	tests.InitHelpers(t)
	defer swapclient(10)()
	_, err := orderClient.get("/", nil)
	tests.Exp(err)
	_, err = orderClient.get("/invalid path", nil)
	tests.Exp(err)
	b, err := orderClient.post("/invalid path", nil, bytes.NewReader(make([]byte, 1)))
	tests.Exp(err)
	if len(b) != 0 {
		t.Error("expected zero length response")
	}
	_, err = orderClient.post("invalid path", nil, bytes.NewReader(nil))
	tests.Exp(err)
	_, err = orderClient.post("/power/price-order", nil, bytes.NewReader([]byte{}))
	tests.Exp(err)
	cli := &client{
		Client: &http.Client{
			Transport: &http.Transport{
				DialTLS: func(string, string) (net.Conn, error) {
					return nil, errors.New("stop")
				},
			},
			Timeout: time.Second,
		},
	}
	resp, err := cli.get("/power/store/4336/profile", nil)
	tests.Exp(err)
	if resp != nil {
		t.Error("should not have gotten any response data")
	}
	b, err = cli.post("/invalid path", nil, bytes.NewReader(make([]byte, 1)))
	tests.Exp(err)
	if b != nil {
		t.Error("expected zero length response")
	}
	req, err := http.NewRequest("GET", "https://www.google.com/", nil)
	tests.Check(err)
	resp, err = orderClient.do(req)
	tests.Exp(err, "expected an error because we found an html page\n")
	if err == nil {
		t.Error("expected an error because we found an html page")
	}
	req, err = http.NewRequest("GET", "https://hjfkghfdjkhgfjkdhgjkdghfdjk.com", nil)
	tests.Check(err)
	resp, err = orderClient.do(req)
	tests.Exp(err)
}

func TestDominosErrors(t *testing.T) {
	order := &Order{
		LanguageCode:  "en",
		ServiceMethod: "Delivery",
		Products: []*OrderProduct{
			{
				ItemCommon: ItemCommon{Code: "12SCREEN"},
				Opts: map[string]interface{}{
					"C": map[string]string{"1/1": "1"},
					"P": map[string]string{"1/1": "1.5"},
				},
				Qty: 1,
			},
		},
		StoreID: "4336",
		OrderID: "",
		Address: testAddress(),
	}
	resp, err := orderClient.post("/power/price-order", nil, order.raw())
	if err != nil {
		t.Error(err)
	}
	if err := dominosErr(resp); err != nil && IsFailure(err) {
		t.Error(err)
	}
}

func TestDominosErrorInit(t *testing.T) {
	err := dominosErr([]byte("bad data"))
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("got wrong error type: %T\n", err)
	}
}

func TestDominosErrorFailure(t *testing.T) {
	e := dominosErr([]byte(`
{
	"Status":-1,
	"StatusItems": [{"Code":"Failure","Message":"test msg"}],
	"Order": {"Status": -1,
		"StatusItems": [
			{"Code":"Failure","Message":"test order msg"},
			{"Code":"SomeOtherCode"},
			{"PulseCode": 1, "PulseText": "this isn't the real error format"}
		]}}`))
	if e == nil {
		t.Error("dominos error should not be nil")
	}
	expected := `Dominos Failure (-1)
    Failure Code: 'Failure':
        test order msg
    Failure Code: 'SomeOtherCode'
        PulseCode 1: this isn't the real error format`
	if e.Error() != expected {
		t.Errorf("\nexpected:\n'%s'\ngot:\n'%s'\n", expected, e.Error())
	}
	if len(e.Error()) < 5 {
		t.Error("the error message here seems too small:\n", e.Error())
	}
	dErr := e.(*DominosError)
	if IsOk(dErr) {
		t.Error("no... its not ok!")
	}
	if IsWarning(dErr) {
		t.Error("error is not a warning")
	}
	if !IsFailure(dErr) {
		t.Error("should be a failure")
	}
	if IsWarning(nil) {
		t.Error("nil should not be interpreted as an error")
	}
	if !IsOk(nil) {
		t.Error("IsOk(nil) should be true because a nil error is ok")
	}
	if IsFailure(nil) {
		t.Error("nil is not a failure")
	}
}

func TestValidateCard(t *testing.T) {
	tests.InitHelpers(t)
	tsts := []struct {
		c     Card
		valid bool
	}{
		{NewCard("", "0125", 123), false},
		{NewCard("", "01/25", 123), false},
		{NewCard("370218180742397", "0123", 123), true},
		{NewCard("370218180742397", "01/23", 123), true},
		{NewCard("370218180742397", "1/23", 123), true},
		{NewCard("370218180742397", "01/02", 123), true},
		{NewCard("370218180742397", "13/21", 123), false},
		{NewCard("370218180742397", "0/21", 123), false},
	}

	for _, tc := range tsts {
		if tc.valid {
			if tc.c == nil {
				t.Error("got nil card when it should be valid")
				continue
			}
			tests.Check(ValidateCard(tc.c))
		} else {
			tests.Exp(ValidateCard(tc.c), "expected an error:", tc.c, tc.c.ExpiresOn())
		}
	}
}

func TestParseDate(t *testing.T) {
	tst := []struct {
		s    string
		m, y int
	}{
		{"01/25", 1, 2025},
		{"0125", 1, 2025},
		{"01/2025", 1, 2025},
		{"1/25", 1, 2025},
		{"1/2025", 1, 2025},
		{"012025", -1, -1}, // failure case
		{"11/02", 11, 2002},
		{"11/2002", 11, 2002},
		{"11/2", 11, 202}, // failure case
	}
	var m, y int
	for _, tc := range tst {

		m, y = parseDate(tc.s)
		if m != tc.m {
			t.Errorf("got the wrong month; want %d, got %d", tc.m, m)
		}
		if y != tc.y {
			t.Errorf("got the wrong year; wand %d, got %d", tc.y, y)
		}
	}
}

func TestErrPair(t *testing.T) {
	tt := []struct {
		err error
		exp string
	}{
		{errpair(errors.New("one"), errors.New("two")), "error 1. one\nerror 2. two"},
		{errpair(errors.New("one"), nil), "one"},
		{errpair(nil, errors.New("two")), "two"},
	}
	for i, tc := range tt {
		if tc.err.Error() != tc.exp {
			t.Errorf("test case %d for errpair gave wrong result", i)
		}
	}
	err := errpair(nil, nil)
	if err != nil {
		t.Error("a pair of nil errors should result in one nil error")
	}
}

var (
	testStore *Store
	testMenu  *Menu
)

func testingStore() *Store {
	var (
		service string
		err     error
	)

	if rand.Intn(2) == 1 {
		service = "Carryout"
	} else {
		service = "Delivery"
	}
	if testStore == nil {
		testStore, err = NearestStore(testAddress(), service)
		if err != nil {
			panic(err)
		}
	}
	return testStore
}

func testingMenu() *Menu {
	var err error
	if testMenu == nil {
		testMenu, err = testingStore().Menu()
		if err != nil {
			panic(err)
		}
	}
	return testMenu
}

func storeLocatorHandlerFunc(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		addr := testAddress()
		if q.Get("c") != fmt.Sprintf("%s, %s %s", addr.City(), addr.StateCode(), addr.Zip()) {
			t.Error("bad url query: \"c\"")
		}
		if q.Get("s") != addr.LineOne() {
			t.Error("bad url query: \"s\"")
		}
		fileHandleFunc(t, "./testdata/store-locator.json")(w, r)
	}
}

func storeProfileHandlerFunc(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("not a get req")
			w.WriteHeader(500)
			return
		}
		fileHandleFunc(t, "./testdata/store.json")(w, r)
	}
}
