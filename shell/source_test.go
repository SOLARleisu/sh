// Copyright (c) 2018, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package shell

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/syntax"

	"github.com/kr/pretty"
)

var mapTests = []struct {
	in   string
	want map[string]expand.Variable
}{
	{
		"a=x; b=y",
		map[string]expand.Variable{
			"a": {Value: "x"},
			"b": {Value: "y"},
		},
	},
	{
		"a=x; a=y; X=(a b c)",
		map[string]expand.Variable{
			"a": {Value: "y"},
			"X": {Value: []string{"a", "b", "c"}},
		},
	},
	{
		"a=$(echo foo | sed 's/o/a/g')",
		map[string]expand.Variable{
			"a": {Value: "faa"},
		},
	},
}

var errTests = []struct {
	in   string
	want string
}{
	{
		"rm -rf /",
		"not in whitelist: rm",
	},
	{
		"cat secret >some-file",
		"cannot open path",
	},
}

func TestSource(t *testing.T) {
	for i := range mapTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tc := mapTests[i]
			t.Parallel()
			p := syntax.NewParser()
			file, err := p.Parse(strings.NewReader(tc.in), "")
			if err != nil {
				t.Fatal(err)
			}
			got, err := SourceNode(file)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatal(strings.Join(pretty.Diff(tc.want, got), "\n"))
			}
		})
	}
}

func TestSourceErr(t *testing.T) {
	for i := range errTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tc := errTests[i]
			t.Parallel()
			p := syntax.NewParser()
			file, err := p.Parse(strings.NewReader(tc.in), "")
			if err != nil {
				t.Fatal(err)
			}
			_, err = SourceNode(file)
			if err == nil {
				t.Fatal("wanted non-nil error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not match %q", err, tc.want)
			}
		})
	}
}
