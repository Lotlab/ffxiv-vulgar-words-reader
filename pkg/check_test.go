package pkg_test

import (
	"io"
	"os"
	"testing"

	"github.com/lotlab/ffxiv-vulgar-words-reader/pkg"
)

func loadDict() (*pkg.Dict, error) {
	const testData = "../vulgarwordsfilter_party.dic"

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

type pos struct {
	Begin int
	End   int
}

func TestCheck2(t *testing.T) {
	dict, err := loadDict()
	if err != nil {
		t.Fatal(err)
	}

	testData := map[string]pos{
		"练习队":     {1, 2},
		"魔大陆":     {1, 3},
		"阿拉米格解放军": {4, 7},
		"雇员探险":    {-1, -1},
		"大国防联军":   {-1, -1},
		"2.5sGCD": {4, 7},
		"妖怪手表":    {-1, -1},
		"USA蹦":    {0, 3},
		"马 克思":    {0, 4},
		"版本5.35":  {2, 6},
		"640HQ":   {0, 2},
	}

	for key, val := range testData {
		beg, end, err := dict.CheckString(key)
		if err != nil {
			t.Fatal(err)
		}
		if val.Begin != beg || val.End != end {
			t.Errorf("%s: {%d:%d} not match {%d:%d}", key, val.Begin, val.End, beg, end)
		}
	}
}
