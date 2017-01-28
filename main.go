package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/toqueteos/webbrowser"
	"github.com/abbot/go-http-auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	doc  *document
	lock sync.RWMutex

	documentPath = flag.String("f", "api-spec.yaml", "the full path to the document being edited")
	backendPort  = flag.String("p", "8765", "port for editor's http backend")
	editorPath   = flag.String("se", "builtin", "the full path to swagger-editor installation")
	authName     = flag.String("u", "admin", "username for web authentication")
	authPass     = flag.String("k", "admin", "password for web authentication")
)

type document struct {
	sync.RWMutex
	path  string
	saved bool
	buf   *bytes.Buffer
}

func init() {
	flag.Parse()
	doc = &document{
		buf:  &bytes.Buffer{},
		path: *documentPath,
	}

	go doc.doSync()
}

func (doc *document) doSync() {
	err := doc.open()
	if err != nil {
		log.Println(err)
		return
	}

	tick := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-tick:
			if !doc.saved {
				doc.save()
			}
		}
	}
}

func (doc *document) open() error {
	doc.Lock()
	f, err := os.Open(doc.path)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(doc.path)
		}

		if err != nil {
			return err
		}

	}

	defer f.Close()
	defer doc.Unlock()

	io.Copy(doc.buf, f)
	return nil
}

func (doc *document) save() error {
	doc.RLock()
	f, err := os.Create(doc.path)
	if err != nil {
		return err
	}

	defer f.Close()

	n, err := f.Write(doc.buf.Bytes())
	if err != nil {
		return err
	}
	
	g, err2 := os.Create(doc.path + "_" + time.Now().Format("201701010101"))
        if err2 != nil {
                return err2
        }
        defer g.Close()
        g.Write(doc.buf.Bytes())

	doc.saved = true
	doc.RUnlock()

	log.Printf("%v bytes saved\n", n)
	return nil
}

func handleBackend(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	switch r.Method {
	case http.MethodGet:
		doc.RLock()
		_, err := w.Write(doc.buf.Bytes())
		if err != nil {
			log.Println(err)
		}
		doc.RUnlock()

	case http.MethodPut:
		doc.Lock()
		doc.buf.Reset()
		_, err := io.Copy(doc.buf, r.Body)
		if err != nil {
			log.Println(err)
		}
		doc.saved = false
		doc.Unlock()
	}
}

func handleApp(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	data, err := Asset("swagger-editor" + path)
	if err != nil {
		log.Println(path)
		http.Error(w, "resource not found"+path, http.StatusNotFound)
	}

	contentType := http.DetectContentType(data)
	w.Header().Set("Content-Type", contentType)

	if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}

	w.Write(data)
}
	
func handleAuth(user, realm string) string {
        if user == *authName {
                password := []byte(*authPass)
		encrypted, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
		return string(encrypted)
        }
        return ""
}

func main() {
	webbrowser.Open("http://localhost:" + *backendPort)

	authenticator := auth.NewBasicAuthenticator("Swagger Web", handleAuth)
	http.HandleFunc("/backend", authenticator.Wrap(handleBackend))
	if *editorPath == "builtin" {
		http.HandleFunc("/", authenticator.Wrap(handleApp))
	} else {
		http.Handle("/", http.FileServer(http.Dir(*editorPath)))
	}

	http.ListenAndServe(":8765", nil)
}
