package analyzer

import (
	"dup/dump"
	"dup/metas"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestPurgeBbb(t *testing.T) {
	origin := metas.FileMetas{
		{"a/b/c", 12, time.Time{}, "123"},
		{"aaa", 13, time.Time{}, "234"},
	}
	copy := metas.FileMetas{
		{"aaa", 13, time.Time{}, "234"},
		{"bbb", 14, time.Time{}, "345"},
	}
	analysis := Analyze(origin, copy)
	expected := Analysis{
		Purge: []string{"bbb"},
	}
	if !reflect.DeepEqual(expected, analysis) {
		fmt.Printf("expected: %s\ngot: %s\n", dump.Dump(expected), dump.Dump(analysis))
		t.FailNow()
	}
}
