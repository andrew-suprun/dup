package analyzer

import (
	"dup/dump"
	"dup/metas"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestAnalysis(t *testing.T) {
	origin := metas.FileMetas{
		{"a/b/c", 12, time.Time{}, "123"},
		{"1/aaa", 13, time.Time{}, "234"},
		{"2/bbb", 14, time.Time{}, "777"},
		{"qqq", 15, time.Time{}, "15"},
	}
	copy := metas.FileMetas{
		{"1/aaa", 14, time.Time{}, "777"},
		{"2/bbb", 13, time.Time{}, "234"},
		{"xxx", 14, time.Time{}, "345"},
		{"www", 15, time.Time{}, "15"},
		{"eee", 15, time.Time{}, "15"},
	}
	analysis := Analyze(origin, copy)
	expected := Analysis{
		Purge: []string{"xxx", "eee"},
		Move: []Move{
			{From: "2/bbb", To: "2/temp-bbb"},
			{From: "1/aaa", To: "2/bbb"},
			{From: "2/temp-bbb", To: "1/aaa"},
		},
		Copy: []string{"a/b/c"},
	}
	if !reflect.DeepEqual(expected, analysis) {
		fmt.Printf("expected: %s\ngot: %s\n", dump.Dump(expected), dump.Dump(analysis))
		t.FailNow()
	}
}
