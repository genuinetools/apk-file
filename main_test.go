package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

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

func TestParseHTML(t *testing.T) {
	searchResults := `
<table class="pure-table table-striped table-bordered table-condensed" data-toggle="table">
    <tbody>
	<tr>
	    <th>File</th>
	    <th>Package</th>
	    <th>Branch</th>
	    <th>Repository</th>
	    <th>Architecture</th>
	</tr>
	
	<tr>
	    <td>/usr/lib/php7/modules/posix.so</td>
	    <td><a href="/package/edge/testing/armhf/php7-posix">php7-posix</a></td>
	    <td>edge</td>
	    <td>testing</td>
	    <td>armhf</td>
	</tr>                    
    </tbody>
</table>
`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(searchResults))
	files := getFilesInfo(doc)
	if len(files) != 1 {
		t.Fatalf("expected %d, got %d", 1, len(files))
	}

	expectedFile := fileInfo{
		path:   "/usr/lib/php7/modules/posix.so",
		pkg:    "php7-posix",
		branch: "edge",
		repo:   "testing",
		arch:   "armhf"}

	if files[0] != expectedFile {
		t.Fatalf("expected %v, got %v", expectedFile, files[0])
	}

}
