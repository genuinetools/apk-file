package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/PuerkitoBio/goquery"
	"github.com/genuinetools/apk-file/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
)

const (
	alpineContentsSearchURI = "https://pkgs.alpinelinux.org/contents"
)

type fileInfo struct {
	path, pkg, branch, repo, arch string
}

var (
	arch string
	repo string

	debug bool

	validArches = []string{"x86", "x86_64", "armhf", "armv7", "aarch64", "ppc64le", "s390x", "mips64"}
	validRepos  = []string{"main", "community", "testing"}
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "apk-file"
	p.Description = "Search apk package contents via the command line"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&arch, "arch", "", "arch to search for ("+strings.Join(validArches, ", ")+")")
	p.FlagSet.StringVar(&repo, "repo", "", "repository to search in ("+strings.Join(validRepos, ", ")+")")
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if arch != "" && !in(arch, validArches) {
			return fmt.Errorf("%s is not a valid arch", arch)
		}

		if repo != "" && !in(repo, validRepos) {
			return fmt.Errorf("%s is not a valid repo", repo)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		if len(args) < 1 {
			return errors.New("must pass a file to search for")
		}

		f, p := getFileAndPath(args[0])

		query := url.Values{
			"file":   {f},
			"path":   {p},
			"branch": {""},
			"repo":   {repo},
			"arch":   {arch},
		}

		uri := fmt.Sprintf("%s?%s", alpineContentsSearchURI, query.Encode())
		logrus.Debugf("requesting from %s", uri)
		resp, err := http.Get(uri)
		if err != nil {
			logrus.Fatalf("requesting %s failed: %v", uri, err)
		}
		defer resp.Body.Close()
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			logrus.Fatalf("creating document failed: %v", err)
		}

		// create the writer
		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		io.WriteString(w, "FILE\tPACKAGE\tBRANCH\tREPOSITORY\tARCHITECTURE\n")

		files := getFilesInfo(doc)

		for _, f := range files {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				f.path,
				f.pkg,
				f.branch,
				f.repo,
				f.arch)
		}

		w.Flush()

		return nil
	}

	// Run our program.
	p.Run()
}

func getFilesInfo(d *goquery.Document) []fileInfo {
	files := []fileInfo{}
	d.Find(".pure-table tr:not(:first-child)").Each(func(j int, l *goquery.Selection) {
		f := fileInfo{}
		rows := l.Find("td")
		rows.Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				f.path = s.Text()
			case 1:
				f.pkg = s.Text()
			case 2:
				f.branch = s.Text()
			case 3:
				f.repo = s.Text()
			case 4:
				f.arch = s.Text()
			default:
				logrus.Warnf("Unmapped value for column %d with value %s", i, s.Text())
			}
		})
		files = append(files, f)
	})
	return files
}

func getFileAndPath(arg string) (file string, dir string) {
	file = "*" + path.Base(arg) + "*"
	dir = path.Dir(arg)
	if dir != "" && dir != "." {
		dir = "*" + dir
		file = strings.TrimPrefix(file, "*")
	} else {
		dir = ""
	}
	return file, dir
}

func in(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
