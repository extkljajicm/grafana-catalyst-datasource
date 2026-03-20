package backend

import "testing"

func TestStringOrBool_Unmarshal(t *testing.T) {
	var v StringOrBool

	// boolean true
	if err := v.UnmarshalJSON([]byte("true")); err != nil || v.String() != "true" {
		t.Fatalf("bool true -> %q (err=%v), want 'true'", v, err)
	}

	// boolean false
	if err := v.UnmarshalJSON([]byte("false")); err != nil || v.String() != "false" {
		t.Fatalf("bool false -> %q (err=%v), want 'false'", v, err)
	}

	// string YES
	if err := v.UnmarshalJSON([]byte(`"YES"`)); err != nil || v.String() != "true" {
		t.Fatalf("string YES -> %q (err=%v), want 'true'", v, err)
	}

	// string NO
	if err := v.UnmarshalJSON([]byte(`"NO"`)); err != nil || v.String() != "false" {
		t.Fatalf("string NO -> %q (err=%v), want 'false'", v, err)
	}
}

