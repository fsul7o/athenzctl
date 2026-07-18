package edit

import "encoding/json"

// jsonRoundTrip copies fields from src to dst via JSON. Used to project a
// full resource (Domain / Role / Group) into its *Meta counterpart:
// both types share JSON tags for every meta field, so this exactly
// extracts the meta subset without hand-copying dozens of fields.
func jsonRoundTrip(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
