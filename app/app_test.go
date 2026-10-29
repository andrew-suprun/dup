package app

import (
	"testing"
)

type TestScript []ArchiveScript
type ArchiveScript []MessageExchange
type MessageExchange struct {
	Command any
	Events  []any
}

func Test001(t *testing.T) {
	testRun(t, &TestScript{
		{
			{
				Command: "foo",
				Events:  []any{},
			},
		},
		{},
	})
}

func testRun(t *testing.T, script *TestScript) {

}

// func run(commands []chan<- any, events <-chan any, uiEvents chan<- any, quit <-chan struct{}) {
