package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	"github.com/product-img-api-microservice/files"
)

// Files is a handler for reading and writing files
type Files struct {
	log   hclog.Logger
	store files.Storage
}

// NewFiles creates a new File handler
func NewFiles(s files.Storage, l hclog.Logger) *Files {
	return &Files{store: s, log: l}
}

// UploadRest implements the http.Handler interface
func (f *Files) UploadRest(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fn := vars["filename"]

	f.log.Info("Handle POST", "id", id, "filename", fn)

	// no need to check for invalid id or filename as the mux router will not send requests
	// here unless they have the correct parameters

	f.saveFile(id, fn, rw, r.Body)
}

// UploadMultiPart implements the http.Handler interface
// multipart for is saved In key-value pairs
func (f *Files) UploadMultiPart(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(128*1024)	// Max Size for File sent check
	if err != nil {
		f.log.Error("Bad Request", "error", err)
		http.Error(rw, "Expected multipart form data", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	f.log.Info("id Logged", id)

	if err != nil {
		f.log.Error("Bad Request", "error", err)
		http.Error(rw, "Expected integer id", http.StatusBadRequest)
		return
	}

	file, fileHeaders, err := r.FormFile("file")
	if err != nil {
		f.log.Error("Bad Request", "error", err)
		http.Error(rw, "Expected File", http.StatusBadRequest)
		return
	}

	f.saveFile(r.FormValue("id"), fileHeaders.Filename, rw, file)
}

func (f *Files) invalidURI(uri string, rw http.ResponseWriter) {
	f.log.Error("Invalid path", "path", uri)
	http.Error(rw, "Invalid file path should be in the format: /[id]/[filepath]", http.StatusBadRequest)
}

// saveFile saves the contents of the request to a file
func (f *Files) saveFile(id, path string, rw http.ResponseWriter, r io.Reader) {
	f.log.Info("Save file for product", "id", id, "path", path)

	fp := filepath.Join(id, path)
	err := f.store.Save(fp, r)
	if err != nil {
		f.log.Error("Unable to save file", "error", err)
		http.Error(rw, "Unable to save file", http.StatusInternalServerError)
	}
}
