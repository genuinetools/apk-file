package main

import (
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
	"github.com/sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `             _          __ _ _
  __ _ _ __ | | __     / _(_) | ___
 / _` + "`" + ` | '_ \| |/ /____| |_| | |/ _ \
| (_| | |_) |   <_____|  _| | |  __/
 \__,_| .__/|_|\_\    |_| |_|_|\___|
      |_|

 Search apk package contents via the command line.
 Version: %s

`
	alpineContentsSearchURI = "https://pkgs.alpinelinux.org/contents"
)

type fileInfo struct {
	path, pkg, branch, repo, arch string
}

var (
	arch string
	repo string

	debug bool
	vrsn  bool

	validArches = []string{"x86", "x86_64", "armhf"}
	validRepos  = []string{"main", "community", "testing"}
)

func init() {
	// Parse flags
	flag.StringVar(&arch, "arch", "", "arch to search for ("+strings.Join(validArches, ", ")+")")
	flag.StringVar(&repo, "repo", "", "repository to search in ("+strings.Join(validRepos, ", ")+")")

	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vrsn {
		fmt.Printf("apk-file version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	// Set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if arch != "" && !stringInSlice(arch, validArches) {
		logrus.Fatalf("%s is not a valid arch", arch)
	}

	if repo != "" && !stringInSlice(repo, validRepos) {
		logrus.Fatalf("%s is not a valid repo", repo)
	}
}

func main() {
	if flag.NArg() < 1 {
		logrus.Fatal("must pass a file to search for.")
	}

	f, p := getFileAndPath(flag.Arg(0))

	query := url.Values{
		"file":   {f},
		"path":   {p},
		"branch": {""},
		"repo":   {repo},
		"arch":   {arch},
	}

	uri := fmt.Sprintf("%s?%s", alpineContentsSearchURI, query.Encode())
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
				logrus.Warn("Unmapped value for column %d with value %s", i, s.Text())
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
