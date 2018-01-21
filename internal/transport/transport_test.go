package transport_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cloudflare/unsee/internal/mock"
	"github.com/cloudflare/unsee/internal/transport"

	log "github.com/sirupsen/logrus"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type transportTest struct {
	uri     string
	timeout time.Duration
	failed  bool
}

var transportTests = []transportTest{
	transportTest{
		uri: "http://localhost/status",
	},
	transportTest{
		uri:    "http://localhost/404",
		failed: true,
	},
	transportTest{
		uri:    "http://localhost/invalid",
		failed: true,
	},
	transportTest{
		uri: "https://localhost/status",
	},
	transportTest{
		uri:    "https://localhost/404",
		failed: true,
	},
	transportTest{
		uri:    "https://localhost/invalid",
		failed: true,
	},
	transportTest{
		uri: fmt.Sprintf("file://%s", mock.GetAbsoluteMockPath("status", mock.ListAllMocks()[0])),
	},
	transportTest{
		uri:    "file:///non-existing-file.abcdef",
		failed: true,
	},
	transportTest{
		uri:    "file://transport.go",
		failed: true,
	},
}

type mockStatus struct {
	status  string
	integer int
	yes     bool
	no      bool
}

func TestTransport(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	mockJSON := `{
			"response": "success",
			"integer": 123,
			"yes": true,
			"no": false
		}`
	httpmock.RegisterResponder("GET", "http://localhost/status", httpmock.NewStringResponder(200, mockJSON))
	httpmock.RegisterResponder("GET", "http://localhost/404", httpmock.NewStringResponder(404, "404"))
	httpmock.RegisterResponder("GET", "http://localhost/invalid", httpmock.NewStringResponder(200, "bad json}{}"))
	httpmock.RegisterResponder("GET", "https://localhost/status", httpmock.NewStringResponder(200, mockJSON))
	httpmock.RegisterResponder("GET", "https://localhost/404", httpmock.NewStringResponder(404, "404"))
	httpmock.RegisterResponder("GET", "https://localhost/invalid", httpmock.NewStringResponder(200, "bad json}{}"))

	for _, testCase := range transportTests {
		tr, err := transport.NewTransport(testCase.uri, testCase.timeout)
		if err != nil {
			t.Error(err)
		}

		source, err := tr.Read(testCase.uri)
		if err != nil {
			if !testCase.failed {
				t.Errorf("[%s] transport Read() failed with: %s", testCase.uri, err)
			}
			continue
		}

		r := mockStatus{}
		err = json.NewDecoder(source).Decode(&r)
		source.Close()

		if (err != nil) != testCase.failed {
			t.Errorf("[%s] Expected failure: %v, Read() failed: %v, error: %s", testCase.uri, testCase.failed, (err != nil), err)
		}
	}
}
