package analyzer

import (
	"dup/dump"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestAnalysis(t *testing.T) {
	now := time.Now()
	var nilFolder *Folder

	origin := nilFolder.AddSubfolder("-origin-")
	a := origin.AddSubfolder("a")
	b := a.AddSubfolder("b")
	c := b.AddSubfolder("c")
	c.AddFile("d.d", 1, now, "d.d-hash")
	c.Id = "ID"

	copy := nilFolder.AddSubfolder("-copy-")
	e := copy.AddSubfolder("e")
	e.AddFile("d.d", 1, now, "d.d-hash")
	e.Id = "ID"
	fmt.Println(dump.Dump(copy))

	analysis := Analyze(origin, copy, "tmp")
	expected := &Analysis{
		Move: []Move{
			{From: "e", To: "a/b/c"},
		},
	}
	if !reflect.DeepEqual(expected, analysis) {
		fmt.Printf("expected: %s\ngot: %s\n", dump.Dump(expected), dump.Dump(analysis))
		t.FailNow()
	}
}
