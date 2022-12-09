package pkg_test

import "testing"

func TestDump(t *testing.T) {
	dict, err := loadDict()
	if err != nil {
		t.Fatal(err)
	}

	lines, err := dict.DumpDict()
	if err != nil {
		t.Fatal(err)
	}

	if len(lines) < 10000 {
		t.Fail()
	}
}
