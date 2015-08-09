// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package godoc is a stripped-down version of
// golang.org/x/tools/godoc/cmd/godoc for an experimental godoc mobile app.
package godoc

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"sync"
	"text/template"

	"golang.org/x/mobile/asset"
	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/redirect"
	"golang.org/x/tools/godoc/static"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

// TODO: Instead of zipfs, demonstrate how to run a http fileserver with files in /assets dir?
// TODO: Put GOROOT in external storage
// TODO: Download go src from the network instead of using zip file.
// TODO: Support packages in GOPATH

// loadFS cannot be called from init due to https://golang.org/issues/12077.
var once sync.Once

func initOnce() {
	once.Do(func() {
		loadFS()
		registerHandlers()
	})
}

// zip file created with github.com/hyangah/godoc/zip.bash
const gorootZipFile = "go.zip"

// loadFS reads the zip file of Go source code and binds.
func loadFS() {
	// Load GOROOT (in zip file)
	rsc, err := asset.Open(gorootZipFile) // asset file, never closed.
	if err != nil {
		panic(err)
	}
	offset, err := rsc.Seek(0, os.SEEK_END)
	if err != nil {
		panic(err)
	}
	if _, err = rsc.Seek(0, os.SEEK_SET); err != nil {
		panic(err)
	}
	r, err := zip.NewReader(&readerAt{wrapped: rsc}, offset)
	if err != nil {
		panic(err)
	}

	fs.Bind("/", newZipFS(r, gorootZipFile), "/go", vfs.BindReplace)

	// static files for godoc.
	fs.Bind("/lib/godoc", mapfs.New(static.Files), "/", vfs.BindReplace)
}

type readerAt struct {
	mu      sync.Mutex
	wrapped io.ReadSeeker
}

func (r *readerAt) ReadAt(data []byte, off int64) (cnt int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	old, err := r.wrapped.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}
	defer func() {
		if _, err2 := r.wrapped.Seek(old, os.SEEK_SET); err == nil {
			err = err2
		}
	}()

	if _, err = r.wrapped.Seek(off, os.SEEK_SET); err != nil {
		return 0, err
	}

	return r.wrapped.Read(data)
}

func (r *readerAt) Read(data []byte) (cnt int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.wrapped.Read(data)
}

func (r *readerAt) Seek(off int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.wrapped.Seek(off, whence)
}

var (
	corpus *godoc.Corpus
	pres   *godoc.Presentation
	fs     = vfs.NameSpace{}
)

const tabWidth = 4

func registerHandlers() {
	corpus = godoc.NewCorpus(fs)
	corpus.MaxResults = 100
	if err := corpus.Init(); err != nil {
		panic(err)
	}

	/*
		// TODO: enable indexing and analysis.
		// Currently indexing consumes a lot of memory that mobile devices cannot tolerate.
		// Static analysis does not handle the current unusual GOROOT setup.
		corpus.IndexEnabled = true
		corpus.IndexDirectory = func(dir string) bool {
			return dir != "/pkg" && !strings.HasPrefix(dir, "/pkg/")
		}
		corpus.IndexInterval = -1
		go corpus.RunIndexer()
		go analysis.Run(true, &corpus.Analysis)
	*/

	// presentation
	pres = godoc.NewPresentation(corpus)
	pres.TabWidth = tabWidth
	pres.ShowTimestamps = false
	pres.ShowExamples = true
	pres.DeclLinks = true
	pres.HTMLMode = true
	pres.PackageText = readTemplate("package.txt")
	pres.SearchText = readTemplate("search.txt")
	pres.CallGraphHTML = readTemplate("callgraph.html")
	pres.DirlistHTML = readTemplate("dirlist.html")
	pres.ErrorHTML = readTemplate("error.html")
	pres.ExampleHTML = readTemplate("example.html")
	pres.GodocHTML = readTemplate("godoc.html")
	pres.ImplementsHTML = readTemplate("implements.html")
	pres.MethodSetHTML = readTemplate("methodset.html")
	pres.PackageHTML = readTemplate("package.html")
	pres.SearchHTML = readTemplate("search.html")
	pres.SearchDocHTML = readTemplate("searchdoc.html")
	pres.SearchCodeHTML = readTemplate("searchcode.html")
	pres.SearchTxtHTML = readTemplate("searchtxt.html")
	pres.SearchDescXML = readTemplate("opensearch.xml")

	// handlers
	http.Handle("/doc/play/", pres.FileServer())
	http.Handle("/robots.txt", pres.FileServer()) // do we care?
	http.Handle("/", pres)
	http.Handle("/pkg/C/", redirect.Handler("/cmd/cgo/"))

	redirect.Register(nil)
}

func readTemplate(name string) *template.Template {
	if pres == nil {
		panic("no global Presentation set yet")
	}
	path := "lib/godoc/" + name

	// use underlying file system fs to read the template file
	// (cannot use template ParseFile functions directly)
	data, err := vfs.ReadFile(fs, path)
	if err != nil {
		panic(err)
	}
	// be explicit with errors (for app engine use)
	t, err := template.New(name).Funcs(pres.FuncMap()).Parse(string(data))
	if err != nil {
		panic(err)
	}
	return t
}
