package describe

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AthenZ/athenz/clients/go/zms"
	rdl "github.com/ardielle/ardielle-go/rdl"

	"github.com/fsul7o/athenzctl/internal/printer"
)

func TestDescribeDomainUsesDomainData(t *testing.T) {
	domain := zms.NewDomainData()
	domain.Name = "example.com"
	domain.Description = "example domain"
	domain.Modified = rdl.Timestamp{Time: time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)}
	domain.Policies.Signature = "policy-signature"
	domain.Policies.KeyId = "policy-key"
	domain.Policies.Contents.Domain = domain.Name

	response, err := json.Marshal(&zms.SignedDomains{
		Domains: []*zms.SignedDomain{{Domain: domain}},
	})
	if err != nil {
		t.Fatal(err)
	}

	client := zms.NewClient("https://zms.example.test", roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/sys/modified_domains" {
			t.Errorf("path = %q, want /sys/modified_domains", r.URL.Path)
		}
		for key, want := range map[string]string{
			"domain":     "example.com",
			"metaonly":   "false",
			"metaattr":   "all",
			"master":     "true",
			"conditions": "true",
		} {
			if got := r.URL.Query().Get(key); got != want {
				t.Errorf("query %s = %q, want %q", key, got, want)
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(response)),
			Request:    r,
		}, nil
	}))

	var jsonOutput bytes.Buffer
	if err := describeDomain(&jsonOutput, &client, "example.com", printer.FormatJSON); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"\"name\": \"example.com\"", "\"roles\":", "\"groups\":", "\"policies\":", "\"services\":", "\"entities\":"} {
		if !strings.Contains(jsonOutput.String(), want) {
			t.Errorf("JSON output = %q, want %q", jsonOutput.String(), want)
		}
	}

	var prettyOutput bytes.Buffer
	if err := describeDomain(&prettyOutput, &client, "example.com", printer.FormatTable); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Name:", "example.com", "Roles:", "Groups:", "Policies:"} {
		if !strings.Contains(prettyOutput.String(), want) {
			t.Errorf("pretty output = %q, want %q", prettyOutput.String(), want)
		}
	}
}

func TestDescribeDomainReturnsNotFoundForEmptyResponse(t *testing.T) {
	client := zms.NewClient("https://zms.example.test", roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"domains":[]}`)),
			Request:    r,
		}, nil
	}))
	var output bytes.Buffer
	err := describeDomain(&output, &client, "missing.example", printer.FormatJSON)
	if err == nil || !strings.Contains(err.Error(), "domain with name missing.example wasn't found") {
		t.Fatalf("error = %v, want not-found error", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
