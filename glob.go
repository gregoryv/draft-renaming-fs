// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"path"
	"runtime"
)

// A GlobSys is a file system with a Glob method.
type GlobSys interface {
	System

	// Glob returns the names of all files matching pattern,
	// providing an implementation of the top-level
	// Glob function.
	Glob(pattern string) ([]string, error)
}

// Glob returns the names of all files matching pattern or nil
// if there is no matching file. The syntax of patterns is the same
// as in Match. The pattern may describe hierarchical names such as
// /usr/*/bin/ed (assuming the Separator is '/').
//
// Glob ignores file system errors such as I/O errors reading directories.
// The only possible returned error is path.ErrBadPattern, when pattern
// is malformed.
//
// If System implements GlobFS, Glob calls fs.Glob.
// Otherwise, Glob uses ReadDir to traverse the directory tree
// and look for matches for the pattern.
func Glob(sys System, pattern string) (matches []string, err error) {
	if sys, ok := sys.(GlobSys); ok {
		return sys.Glob(pattern)
	}

	if !hasMeta(pattern) {
		if _, err = Stat(sys, pattern); err != nil {
			return nil, nil
		}
		return []string{pattern}, nil
	}

	dir, file := path.Split(pattern)
	dir = cleanGlobPath(dir)

	if !hasMeta(dir) {
		return glob(sys, dir, file, nil)
	}

	// Prevent infinite recursion. See issue 15879.
	if dir == pattern {
		return nil, path.ErrBadPattern
	}

	var m []string
	m, err = Glob(sys, dir)
	if err != nil {
		return
	}
	for _, d := range m {
		matches, err = glob(sys, d, file, matches)
		if err != nil {
			return
		}
	}
	return
}

// cleanGlobPath prepares path for glob matching.
func cleanGlobPath(path string) string {
	switch path {
	case "":
		return "."
	default:
		return path[0 : len(path)-1] // chop off trailing separator
	}
}

// glob searches for files matching pattern in the directory dir
// and appends them to matches. If the directory cannot be
// opened, it returns the existing matches. New matches are
// added in lexicographical order.
func glob(sys System, dir, pattern string, matches []string) (m []string, e error) {
	m = matches
	infos, err := ReadDir(sys, dir)
	if err != nil {
		return // ignore I/O error
	}

	for _, info := range infos {
		n := info.Name()
		matched, err := path.Match(pattern, n)
		if err != nil {
			return m, err
		}
		if matched {
			m = append(m, path.Join(dir, n))
		}
	}
	return
}

// hasMeta reports whether path contains any of the magic characters
// recognized by Match.
func hasMeta(path string) bool {
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == '*' || c == '?' || c == '[' || runtime.GOOS == "windows" && c == '\\' {
			return true
		}
	}
	return false
}
