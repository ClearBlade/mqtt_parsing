package mqtt_parsing

import (
	"fmt"
	"testing"
)

func compareStrArys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Test_SlashAgnosticPlacement(t *testing.T) {
	pl := func(s string, tp TopicPath) string {
		return fmt.Sprintf("Didn't parse %s properly %+v properly", s, tp)
	}

	fba := []string{"foo", "bar"}
	fbba := []string{"foo", "bar", "baz"}
	fa := []string{"foo"}
	if compareStrArys(fbba, fa) {
		t.Error("said two differeing slices were the same")
	}
	if !compareStrArys(fbba, fbba) {
		t.Error("said two same slices were different")
	}
	tp, valid := NewTopicPath("/foo/bar")
	if !compareStrArys(tp.Split, fba) || !valid {
		t.Error(pl("foo/bar", tp))
	}
	tp2, valid := NewTopicPath("foo/bar")
	if !compareStrArys(tp2.Split, fba) || !valid {
		t.Error(pl("foo/bar", tp2))
	}
	tp3, valid := NewTopicPath("/foo/bar")
	if !compareStrArys(tp3.Split, fba) || !valid {
		t.Error(pl("/foo/bar", tp3))
	}

	tp4, valid := NewTopicPath("foo")
	if !compareStrArys(tp4.Split, fa) || !valid {
		t.Error(pl("foo", tp4))
	}
	tp5, valid := NewTopicPath("/foo")
	if !compareStrArys(tp5.Split, fa) || !valid {
		t.Error(pl("foo", tp5))
	}
	tp6, valid := NewTopicPath("foo/")
	if !compareStrArys(tp6.Split, fa) || !valid {
		t.Error(pl("foo", tp6))
	}

	tp7, valid := NewTopicPath("/foo/bar/baz")
	if !compareStrArys(tp7.Split, fbba) || !valid {
		t.Error(pl("/foo/bar/baz", tp7))
	}
	tp8, valid := NewTopicPath("foo/bar/baz")
	if !compareStrArys(tp8.Split, fbba) || !valid {
		t.Error(pl("foo/bar/baz", tp8))
	}
}

func Test_NewTopicPath(t *testing.T) {
	//test to check valid
	tp, valid := NewTopicPath("/foo/bar/baz")
	if len(tp.Whole) < 1 {
		t.Error("rejecting wildcard data")
	}
	if tp.Wildcard {
		t.Error("Falsely identifying wildcard")
	}

	//testing to check inwildcard
	tp, valid = NewTopicPath("/foo/#/baz")
	if valid {
		t.Error("accepting inwildcard topic /foo/#/baz")
	}

	tp, valid = NewTopicPath("/foo/+")
	if !tp.Wildcard {
		t.Error("not identifying plus wildcard")
	}

	tp, valid = NewTopicPath("/foo/+/+")
	if !tp.Wildcard {
		t.Error("Not identifying /foo/+/+ as wildcard")
	}

	tp, valid = NewTopicPath("/#")

	if !tp.Wildcard {
		t.Error("identifying wildcard as not wild")
	}
	if len(tp.Split) != 1 {
		t.Error("this needs to be a length of 1")
	}

	tp, valid = NewTopicPath("foo/bar/#")
	if len(tp.Whole) < 1 {
		t.Error("falsely rejecting foo/bar/#")
	}
	if !valid {
		t.Error("falsely rejecting foo/bar/# topic as not a wildcard")
	}

}

func Test_trimPath(t *testing.T) {
	tp, _ := NewTopicPath("/foo/bar/baz")
	newPath := TrimPath(tp, 2, false)
	if newPath.Whole != "foo/bar" {
		t.Error("invalid path trimming " + newPath.Whole)

	}
}

func Test_TestSomeCharactersInTopic(t *testing.T) {
	_, valid := NewTopicPath("TEST_PUB_ONE")
	if !valid {
		t.Error("invalid topicpath", valid)
	}
	_, valid = NewTopicPath("d01ba5f8-23fd-11e4-a1c6-bc764e0487f9")
	if !valid {
		t.Error("invalid topicpath", valid)
	}
	_, valid = NewTopicPath("/⌣/🜁/Λ")
}
