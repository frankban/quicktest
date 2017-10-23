// Licensed under the MIT license, see LICENCE file for details.

package testerror_test

import (
	"errors"
	"testing"

	"github.com/frankban/quicktest/testerror"
)

func TestBadCheckf(t *testing.T) {
	err := testerror.BadCheckf("bad %s", "wolf")
	expectedMessage := "bad wolf"
	if err.Error() != expectedMessage {
		t.Fatalf("bad check error:\ngot  %q\nwant %q", err, expectedMessage)
	}
}

func TestIsBadCheck(t *testing.T) {
	err := testerror.BadCheckf("bad wolf")
	if !testerror.IsBadCheck(err) {
		t.Fatalf("expected %v to be a bad check", err)
	}
	err = errors.New("bad wolf")
	if testerror.IsBadCheck(err) {
		t.Fatalf("expected %v to not be a bad check", err)
	}
}

func TestMismatch(t *testing.T) {
	err := &testerror.Mismatch{
		Message: "bad wolf",
		Got:     "another pattern",
		Pattern: "my pattern",
	}
	expectedMessage := "bad wolf:\n(-text +pattern)\n\t-: \"another pattern\"\n\t+: \"my pattern\"\n"
	if err.Error() != expectedMessage {
		t.Fatalf("mismatch error:\ngot  %q\nwant %q", err, expectedMessage)
	}
}

func TestNotEqual(t *testing.T) {
	err := &testerror.NotEqual{
		Message: "bad wolf",
		Got:     47,
		Want:    42,
	}
	expectedMessage := "bad wolf:\n(-got +want)\n\t-: 47\n\t+: 42\n"
	if err.Error() != expectedMessage {
		t.Fatalf("not equal error:\ngot  %q\nwant %q", err, expectedMessage)
	}
}
