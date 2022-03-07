package diff

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"strings"
	"text/tabwriter"
)

// StringDiff diffs a and b and returns a human readable diff.
func StringDiff(a, b string) string {
	ba := []byte(a)
	bb := []byte(b)
	out := []byte{}
	i := 0
	for ; i < len(ba) && i < len(bb); i++ {
		if ba[i] != bb[i] {
			break
		}
		out = append(out, ba[i])
	}
	out = append(out, []byte("\n\nA: ")...)
	out = append(out, ba[i:]...)
	out = append(out, []byte("\n\nB: ")...)
	out = append(out, bb[i:]...)
	out = append(out, []byte("\n\n")...)
	return string(out)
}

// ObjectDiff prints the diff of two go objects and fails if the objects
// contain unhandled unexpected fields.
func ObjectDiff(a, b interface{}) string {
	return legacyDiff(a, b)
}

// ObjectGoPrintSideBySide prints a and b as textual dumps side by side,
// enabling easy visual scanning for mismatches.
func ObjectGoPrintSideBySide(a, b interface{}) string {
	s := spew.ConfigState{
		Indent: " ",
		// Extra deep spew.
		DisableMethods: true,
	}
	sA := s.Sdump(a)
	sB := s.Sdump(b)

	linesA := strings.Split(sA, "\n")
	linesB := strings.Split(sB, "\n")
	width := 0
	for _, s := range linesA {
		l := len(s)
		if l > width {
			width = l
		}
	}
	for _, s := range linesB {
		l := len(s)
		if l > width {
			width = l
		}
	}
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, width, 0, 1, ' ', 0)
	max := len(linesA)
	if len(linesB) > max {
		max = len(linesB)
	}
	for i := 0; i < max; i++ {
		var a, b string
		if i < len(linesA) {
			a = linesA[i]
		}
		if i < len(linesB) {
			b = linesB[i]
		}
		fmt.Fprintf(w, "%s\t%s\n", a, b)
	}
	w.Flush()
	return buf.String()
}

func legacyDiff(a, b interface{}) string {
	return cmp.Diff(a, b)
}
