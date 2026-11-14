package analyzer

import (
	"dup/dump"
	"dup/metas"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func Test(t *testing.T) {
	origin := metas.FileMetas{
		{"a/b/c", 12, time.Time{}, "123"},
		{"aaa", 13, time.Time{}, "234"},
		{"bbb", 14, time.Time{}, "777"},
	}
	copy := metas.FileMetas{
		{"aaa", 14, time.Time{}, "777"},
		{"bbb", 13, time.Time{}, "234"},
		{"xxx", 14, time.Time{}, "345"},
	}
	analysis := Analyze(origin, copy)
	expected := Analysis{
		Purge: []string{"xxx"},
		Move: []Move{
			{From: "bbb", To: "temp-bbb"},
			{From: "aaa", To: "bbb"},
			{From: "temp-bbb", To: "aaa"},
		},
		Copy: []string{"a/b/c"},
	}
	if !reflect.DeepEqual(expected, analysis) {
		fmt.Printf("expected: %s\ngot: %s\n", dump.Dump(expected), dump.Dump(analysis))
		t.FailNow()
	}
}
