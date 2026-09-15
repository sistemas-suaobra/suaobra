package server

import "testing"

func TestExtractPhoneDDD(t *testing.T) {
	cases := map[string]string{
		"11999750289":  "11",
		"489737451":    "48",
		"5511999750289": "11",
		"19999893942":  "19",
		"123":          "",
	}
	for in, want := range cases {
		if got := extractPhoneDDD(in); got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
}

func TestPhoneDDDMatchesUF(t *testing.T) {
	if !phoneDDDMatchesUF("11999750289", "SP") {
		t.Fatal("SP+11 should match")
	}
	if !phoneDDDMatchesUF("19999893942", "SP") {
		t.Fatal("SP+19 should match")
	}
	if phoneDDDMatchesUF("489737451", "SP") {
		t.Fatal("SP+48 should NOT match")
	}
	if phoneDDDMatchesUF("11999750289", "DF") {
		t.Fatal("DF+11 should NOT match")
	}
}
