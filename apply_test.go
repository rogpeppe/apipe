package main

import (
	"testing"

	qt "github.com/go-quicktest/qt"
)

func TestApplyDiff(t *testing.T) {
	tests := []struct {
		name      string
		old       string
		new       string
		wantEdits []edit
	}{{
		name:      "no change",
		old:       "hello\nworld\n",
		new:       "hello\nworld\n",
		wantEdits: nil,
	}, {
		name:      "simple change",
		old:       "hello\nold\n",
		new:       "hello\nnew\n",
		wantEdits: []edit{{Addr: "#6,#9", Data: "new"}},
	}, {
		name:      "delete line",
		old:       "hello\nworld\n",
		new:       "hello\n",
		wantEdits: []edit{{Addr: "#6,#12", Data: ""}},
	}, {
		name:      "add line after",
		old:       "hello\n",
		new:       "hello\nworld\n",
		wantEdits: []edit{{Addr: "#6,#6", Data: "world\n"}},
	}, {
		name:      "add line before",
		old:       "hello\n",
		new:       "world\nhello\n",
		wantEdits: []edit{{Addr: "#0,#0", Data: "world\n"}},
	}, {
		name:      "add trailing newline",
		old:       "hello\nworld",
		new:       "hello\nworld\n",
		wantEdits: []edit{{Addr: "#11,#11", Data: "\n"}},
	}, {
		name:      "remove trailing newline",
		old:       "hello\nworld\n",
		new:       "hello\nworld",
		wantEdits: []edit{{Addr: "#11,#12", Data: ""}},
	}, {
		name: "multi-line change",
		old:  "a\nb\nc\nd\n",
		new:  "a\nx\ny\nz\nd\n",
		wantEdits: []edit{
			{Addr: "#2,#3", Data: "x\ny"},
			{Addr: "#6,#7", Data: "z"},
		},
	}, {
		name: "multiple hunks",
		old:  "a\nb\nc\nd\ne\n",
		new:  "a\nB\nc\nD\ne\n",
		wantEdits: []edit{
			{Addr: "#2,#3", Data: "B"},
			{Addr: "#6,#7", Data: "D"},
		},
	}, {
		name:      "replace all",
		old:       "old\n",
		new:       "new\n",
		wantEdits: []edit{{Addr: "#0,#3", Data: "new"}},
	}, {
		name:      "empty to content",
		old:       "",
		new:       "hello\n",
		wantEdits: []edit{{Addr: "#0,#0", Data: "hello\n"}},
	}, {
		name:      "content to empty",
		old:       "hello\n",
		new:       "",
		wantEdits: []edit{{Addr: "#0,#6", Data: ""}},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got []edit
			err := applyDiff([]byte(test.old), []byte(test.new), func(addr string, data []byte) error {
				got = append(got, edit{Addr: addr, Data: string(data)})
				return nil
			})
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.DeepEquals(got, test.wantEdits))
		})
	}
}

type edit struct {
	Addr string
	Data string
}
