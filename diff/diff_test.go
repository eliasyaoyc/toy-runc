package diff

import (
	"fmt"
	"testing"
)

func TestStringDiff(t *testing.T) {
	diff := StringDiff("aaabb", "aaacc")
	expect := "aaa\n\nA: bb\n\nB: cc\n\n"
	if diff != expect {
		t.Errorf("diff returned %v", diff)
	}
}

type A struct {
	name string
}

type B struct {
	name string
}

func TestObjectGoPrintSideBySide(t *testing.T) {
	a := A{
		name: "JeoA",
	}
	b := B{
		name: "JeoB",
	}
	side := ObjectGoPrintSideBySide(a, b)
	fmt.Println(side)
}
