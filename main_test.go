package main

import "testing"

func TestGetFileAndPath(t *testing.T) {
	args := map[string]map[string]string{
		"bin/file": {
			"file": "file*",
			"path": "*bin",
		},
		"posix": {
			"file": "*posix*",
			"path": "",
		},
		"/usr/bin/bash": {
			"file": "bash*",
			"path": "*/usr/bin",
		},
	}

	for input, output := range args {
		f, p := getFileAndPath(input)
		if output["file"] != f {
			t.Fatalf("expected %q, got %q", output["file"], f)
		}
		if output["path"] != p {
			t.Fatalf("expected %q, got %q", output["path"], p)
		}
	}
}
