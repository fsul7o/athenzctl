package lookup

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func TestParseDomainSelector(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		wantMember    zms.ResourceName
		wantRole      zms.ResourceName
		wantTagKey    zms.TagKey
		wantTagValue  zms.TagCompoundValue
		wantAccount   string
		wantAzure     string
		wantGCP       string
		wantProductID string
		wantYPMID     *int32
		wantBusiness  string
	}{
		{
			name:       "role",
			raw:        "member=user.jane,role=admin",
			wantMember: "user.jane",
			wantRole:   "admin",
		},
		{
			name:         "tag with value",
			raw:          "tagKey=team,tagValue=security",
			wantTagKey:   "team",
			wantTagValue: "security",
		},
		{
			name:       "tag without value",
			raw:        "tagKey=team",
			wantTagKey: "team",
		},
		{name: "account", raw: "account=1234567890", wantAccount: "1234567890"},
		{name: "azure", raw: "azure=subscription-id", wantAzure: "subscription-id"},
		{name: "gcp", raw: "gcp=project-id", wantGCP: "project-id"},
		{
			name:      "numeric product id uses ypmid",
			raw:       "productId=10001",
			wantYPMID: int32Ptr(10001),
		},
		{
			name:          "string product id",
			raw:           "productId=product-001",
			wantProductID: "product-001",
		},
		{name: "business service", raw: "businessService=payments", wantBusiness: "payments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDomainSelector(tt.raw)
			if err != nil {
				t.Fatalf("parseDomainSelector(%q) error = %v", tt.raw, err)
			}
			if got.roleMember != tt.wantMember || got.roleName != tt.wantRole {
				t.Fatalf("role selector = (%q, %q), want (%q, %q)", got.roleMember, got.roleName, tt.wantMember, tt.wantRole)
			}
			if got.tagKey != tt.wantTagKey || got.tagValue != tt.wantTagValue {
				t.Fatalf("tag selector = (%q, %q), want (%q, %q)", got.tagKey, got.tagValue, tt.wantTagKey, tt.wantTagValue)
			}
			if got.account != tt.wantAccount || got.azure != tt.wantAzure || got.gcp != tt.wantGCP {
				t.Fatalf("cloud selector = (%q, %q, %q), want (%q, %q, %q)", got.account, got.azure, got.gcp, tt.wantAccount, tt.wantAzure, tt.wantGCP)
			}
			if got.productID != tt.wantProductID || got.businessService != tt.wantBusiness {
				t.Fatalf("product/business selector = (%q, %q), want (%q, %q)", got.productID, got.businessService, tt.wantProductID, tt.wantBusiness)
			}
			if !reflect.DeepEqual(got.productNumber, tt.wantYPMID) {
				t.Fatalf("ypmid = %v, want %v", got.productNumber, tt.wantYPMID)
			}
		})
	}
}

func TestParseDomainSelectorRejectsUnsupportedCombinations(t *testing.T) {
	tests := []string{
		"",
		"member=user.jane",
		"role=admin",
		"tagValue=security",
		"member=user.jane,role=admin,account=123",
		"tagKey=team,account=123",
		"account=123,account=456",
		"account=123,unknown=value",
		"account!=123",
		"account=",
		"account=123,,role=admin",
	}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := parseDomainSelector(raw); err == nil {
				t.Fatalf("parseDomainSelector(%q) succeeded, want error", raw)
			}
		})
	}
}

func TestLookupDomainSendsZMSQueries(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		want     url.Values
	}{
		{
			name:     "role",
			selector: "member=user.jane,role=admin",
			want:     url.Values{"member": {"user.jane"}, "role": {"admin"}},
		},
		{
			name:     "tag",
			selector: "tagKey=team,tagValue=security",
			want:     url.Values{"tagKey": {"team"}, "tagValue": {"security"}},
		},
		{
			name:     "aws account",
			selector: "account=1234567890",
			want:     url.Values{"account": {"1234567890"}},
		},
		{
			name:     "azure subscription",
			selector: "azure=subscription-id",
			want:     url.Values{"azure": {"subscription-id"}},
		},
		{
			name:     "gcp project",
			selector: "gcp=project-id",
			want:     url.Values{"gcp": {"project-id"}},
		},
		{
			name:     "numeric product id",
			selector: "productId=10001",
			want:     url.Values{"ypmid": {"10001"}},
		},
		{
			name:     "string product id",
			selector: "productId=product-001",
			want:     url.Values{"productId": {"product-001"}},
		},
		{
			name:     "business service",
			selector: "businessService=payments",
			want:     url.Values{"businessService": {"payments"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/domain" {
					t.Errorf("request path = %q, want /domain", r.URL.Path)
				}
				if !reflect.DeepEqual(r.URL.Query(), tt.want) {
					t.Errorf("query = %v, want %v", r.URL.Query(), tt.want)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"names":["example.com"]}`))
			}))
			defer server.Close()

			selector, err := parseDomainSelector(tt.selector)
			if err != nil {
				t.Fatal(err)
			}
			client := zms.NewClient(server.URL, nil)
			var out bytes.Buffer
			if err := lookupDomain(&out, &client, selector, printer.FormatJSON); err != nil {
				t.Fatal(err)
			}
			var response zms.DomainList
			if err := json.Unmarshal(out.Bytes(), &response); err != nil {
				t.Fatalf("decode output: %v; output=%s", err, out.String())
			}
			if len(response.Names) != 1 || response.Names[0] != "example.com" {
				t.Fatalf("response = %+v", response)
			}
		})
	}
}

func TestLookupDomainRendersTableAndYAML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"names":["example.com"]}`))
	}))
	defer server.Close()

	selector, err := parseDomainSelector("account=1234567890")
	if err != nil {
		t.Fatal(err)
	}
	client := zms.NewClient(server.URL, nil)

	tests := []struct {
		name   string
		format printer.Format
		want   string
	}{
		{name: "table", format: printer.FormatTable, want: "NAME\nexample.com"},
		{name: "yaml", format: printer.FormatYAML, want: "names:\n  - example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := lookupDomain(&out, &client, selector, tt.format); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), tt.want) {
				t.Fatalf("output = %q, want substring %q", out.String(), tt.want)
			}
		})
	}
}

func TestLookupCommandValidatesBeforeLoadingContext(t *testing.T) {
	cmd := New(&cliopts.Options{})
	cmd.SetArgs([]string{"domain"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--field-selector") {
		t.Fatalf("error = %v, want missing field-selector error", err)
	}

	cmd = New(&cliopts.Options{})
	cmd.SetArgs([]string{"service", "--field-selector", "account=123"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "lookup service") {
		t.Fatalf("error = %v, want unsupported-kind error", err)
	}
}

func int32Ptr(v int32) *int32 { return &v }
