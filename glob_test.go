// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs_test

import (
	. "io/fs"
	"os"
	"testing"
)

var globTests = []struct {
	sys             System
	pattern, result string
}{
	{os.DirFS("."), "glob.go", "glob.go"},
	{os.DirFS("."), "gl?b.go", "glob.go"},
	{os.DirFS("."), "*", "glob.go"},
	{os.DirFS(".."), "*/glob.go", "fs/glob.go"},
}

func TestGlob(t *testing.T) {
	for _, tt := range globTests {
		matches, err := Glob(tt.sys, tt.pattern)
		if err != nil {
			t.Errorf("Glob error for %q: %s", tt.pattern, err)
			continue
		}
		if !contains(matches, tt.result) {
			t.Errorf("Glob(%#q) = %#v want %v", tt.pattern, matches, tt.result)
		}
	}
	for _, pattern := range []string{"no_match", "../*/no_match"} {
		matches, err := Glob(os.DirFS("."), pattern)
		if err != nil {
			t.Errorf("Glob error for %q: %s", pattern, err)
			continue
		}
		if len(matches) != 0 {
			t.Errorf("Glob(%#q) = %#v want []", pattern, matches)
		}
	}
}

func TestGlobError(t *testing.T) {
	_, err := Glob(os.DirFS("."), "[]")
	if err == nil {
		t.Error("expected error for bad pattern; got none")
	}
}

// contains reports whether vector contains the string s.
func contains(vector []string, s string) bool {
	for _, elem := range vector {
		if elem == s {
			return true
		}
	}
	return false
}
