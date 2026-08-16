package get

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func TestGetDomainListFollowsPaginationWhenRequested(t *testing.T) {
	client := zms.NewClient("https://zms.example.test", roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Errorf("limit = %q, want 100", got)
		}
		var body string
		switch r.URL.Query().Get("skip") {
		case "":
			body = `{"names":["example.com"],"next":"cursor-1"}`
		case "cursor-1":
			body = `{"names":["example.org"]}`
		default:
			t.Fatalf("unexpected skip cursor %q", r.URL.Query().Get("skip"))
		}
		return jsonResponse(r, body), nil
	}))

	list, err := getDomainList(&client, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := []zms.DomainName{"example.com", "example.org"}; !sameDomainNames(list.Names, got) {
		t.Fatalf("names = %v, want %v", list.Names, got)
	}
	if list.Next != "" {
		t.Fatalf("next = %q, want empty after --all", list.Next)
	}
}

func TestGetDomainListDoesNotFollowPaginationByDefault(t *testing.T) {
	client := zms.NewClient("https://zms.example.test", roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Errorf("limit = %q, want 100", got)
		}
		return jsonResponse(r, `{"names":["example.com"],"next":"cursor-1"}`), nil
	}))

	list, err := getDomainList(&client, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Names) != 1 || list.Next != "cursor-1" {
		t.Fatalf("list = %+v, want first page with cursor", list)
	}
}

func TestGetDomainRequiresNameForSingularKind(t *testing.T) {
	cmd := New(&cliopts.Options{})
	cmd.SetArgs([]string{"domain"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "use `get domains`") {
		t.Fatalf("error = %v, want singular-domain guidance", err)
	}
}

func TestConfirmAllDomains(t *testing.T) {
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("y\n"))
	cmd.SetErr(&stderr)
	if err := confirmAllDomains(cmd, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "retrieve all domains") {
		t.Fatalf("warning = %q, want all-domain warning", stderr.String())
	}

	cmd.SetIn(strings.NewReader("n\n"))
	if err := confirmAllDomains(cmd, false); err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Fatalf("error = %v, want cancellation", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(r *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}

func sameDomainNames(got, want []zms.DomainName) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
