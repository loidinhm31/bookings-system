package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name                string
	url                 string
	method              string
	params              []postData
	expectedStratusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"major", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"general", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"search", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"reservation", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post-search", "/search-availability", "POST", []postData{
		{key: "start", value: "2022-11-11"},
		{key: "end", value: "2022-11-14"},
	}, http.StatusOK},
	{"post-reservation", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Smith"},
		{key: "email", value: "abc@xyz.com"},
		{key: "phone", value: "333333333"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, test := range theTests {
		if test.method == "GET" {
			resp, err := testServer.Client().Get(testServer.URL + test.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != test.expectedStratusCode {
				t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStratusCode, resp.StatusCode)
			}
		} else if test.method == "POST" {
			values := url.Values{}
			for _, x := range test.params {
				values.Add(x.key, x.value)
			}
			resp, err := testServer.Client().PostForm(testServer.URL+test.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != test.expectedStratusCode {
				t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStratusCode, resp.StatusCode)
			}
		}
	}
}
