package resource

import "testing"

func TestParsePolicyVersion(t *testing.T) {
	cases := []struct {
		in       string
		wantPol  string
		wantVer  string
		wantErr  bool
	}{
		{"read-access", "read-access", "", false},
		{"read-access:v2", "read-access", "v2", false},
		{"  read-access:0  ", "read-access", "0", false},
		{"", "", "", true},
		{":v2", "", "", true},
		{"read-access:", "", "", true},
		{"a:b:c", "", "", true},
	}
	for _, c := range cases {
		got, err := ParsePolicyVersion(c.in)
		if (err != nil) != c.wantErr {
			t.Fatalf("ParsePolicyVersion(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
		}
		if err != nil {
			continue
		}
		if got.Policy != c.wantPol || got.Version != c.wantVer {
			t.Fatalf("ParsePolicyVersion(%q) = %+v, want {%q,%q}", c.in, got, c.wantPol, c.wantVer)
		}
	}
}
