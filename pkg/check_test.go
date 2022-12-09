package pkg_test

import (
	"io"
	"os"
	"testing"

	"github.com/lotlab/ffxiv-vulgar-words-reader/pkg"
)

func loadDict() (*pkg.Dict, error) {
	const testData = "../vulgarwordsfilter.dic"

	fs, err := os.Open(testData)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(fs)
	if err != nil {
		return nil, err
	}

	return pkg.NewDict(bytes)
}

func TestCheck(t *testing.T) {
	dict, err := loadDict()
	if err != nil {
		t.Fatal(err)
	}

	lines, err := dict.DumpDict()
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range lines {
		beg, end, err := dict.CheckString(v)
		if err != nil {
			t.Fatal(err)
		}
		if end < beg {
			t.Errorf("%s is not match!", v)
		}
	}
}

func TestCheck2(t *testing.T) {
	dict, err := loadDict()
	if err != nil {
		t.Fatal(err)
	}

	testData := map[string]bool{
		"练习队":     true,
		"魔大陆":     true,
		"阿拉米格解放军": true,
		"雇员探险":    false,
		"大国防联军":   false,
		"2.5sGCD": true,
		"妖怪手表":    false,
		"USA蹦":    true,
	}

	for key, val := range testData {
		beg, end, err := dict.CheckString(key)
		if err != nil {
			t.Fatal(err)
		}
		if val != (beg < end) {
			t.Fail()
		}
	}
}
