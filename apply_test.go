package main

import (
	"strings"
	"testing"

	qt "github.com/go-quicktest/qt"
)

func TestApply(t *testing.T) {
	tests := []struct {
		name      string
		diffOut   string
		wantEdits []edit
	}{{
		name: "simple change",
		diffOut: `2c2
< old
---
> new
`,
		wantEdits: []edit{{Addr: "2", Data: "new\n"}},
	}, {
		name: "old side no newline",
		diffOut: `2c2
< world
\ No newline at end of file
---
> world
`,
		wantEdits: []edit{{Addr: "2", Data: "world\n"}},
	}, {
		name: "new side no newline",
		diffOut: `2c2
< world
---
> world
\ No newline at end of file
`,
		wantEdits: []edit{{Addr: "2", Data: "world"}},
	}, {
		name: "both sides no newline",
		diffOut: `2c2
< world
\ No newline at end of file
---
> earth
\ No newline at end of file
`,
		wantEdits: []edit{{Addr: "2", Data: "earth"}},
	}, {
		name: "delete with no newline",
		diffOut: `2d1
< world
\ No newline at end of file
`,
		wantEdits: []edit{{Addr: "2", Data: ""}},
	}, {
		name: "add with no newline",
		diffOut: `1a2
> world
\ No newline at end of file
`,
		wantEdits: []edit{{Addr: "1+#0", Data: "world"}},
	}, {
		name: "multi-hunk with no newline on old side",
		diffOut: `1c1
< A
---
> B
3c3
< C
\ No newline at end of file
---
> D
`,
		wantEdits: []edit{
			{Addr: "1", Data: "B\n"},
			{Addr: "3", Data: "D\n"},
		},
	}, {
		name: "multi-line change with no newline on old side",
		diffOut: `2,3c2,3
< b
< c
\ No newline at end of file
---
> x
> c
`,
		wantEdits: []edit{{Addr: "2,3", Data: "x\nc\n"}},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got []edit
			err := apply(strings.NewReader(test.diffOut), func(addr string, data []byte) error {
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
