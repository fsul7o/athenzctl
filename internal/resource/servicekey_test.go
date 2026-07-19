package resource

import "testing"

func TestParseServiceKey(t *testing.T) {
	cases := []struct {
		in      string
		wantSvc string
		wantKey string
		wantErr bool
	}{
		{"myservice", "myservice", "", false},
		{"myservice:v0", "myservice", "v0", false},
		{"  api:key-1  ", "api", "key-1", false},
		{"", "", "", true},
		{":v0", "", "", true},
		{"myservice:", "", "", true},
		{"a:b:c", "", "", true},
	}
	for _, c := range cases {
		got, err := ParseServiceKey(c.in)
		if (err != nil) != c.wantErr {
			t.Fatalf("ParseServiceKey(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
		}
		if err != nil {
			continue
		}
		if got.Service != c.wantSvc || got.KeyID != c.wantKey {
			t.Fatalf("ParseServiceKey(%q) = %+v, want {%q,%q}", c.in, got, c.wantSvc, c.wantKey)
		}
	}
}
