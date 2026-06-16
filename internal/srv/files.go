package srv

import (
	"net/http"
	"os"
	"path"
	"strings"
)

func filesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeErrorHtml(r.Context(), w, http.StatusMethodNotAllowed)
		return
	}

	// Sanity checks
	if !strings.HasPrefix(r.URL.Path, "/") {
		writeErrorHtml(r.Context(), w, http.StatusBadRequest)
		return
	}

	// Redirect trailing '/'
	if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
		w.Header().Set("Location", strings.TrimSuffix(r.URL.Path, "/"))
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	// Redirect index.html
	if strings.HasSuffix(r.URL.Path, "index.html") {
		path := path.Dir(r.URL.Path)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if q := r.URL.RawQuery; q != "" {
			path += "?" + q
		}

		w.Header().Set("Location", path)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	// Don't serve .html files
	if path.Ext(r.URL.Path) == ".html" {
		writeErrorHtml(r.Context(), w, http.StatusNotFound)
		return
	}

	fname := ".dist/app"
	if r.URL.Path == "/" {
		fname += "/index.html"
	} else {
		if path.Dir(r.URL.Path) == "/_a" {
			// Astro (template) assets
			fname = ".dist/tmpl"
		}

		if path.Ext(r.URL.Path) == "" {
			// For svelte, add .html to the path (e.g. `/a` -> `/a.html`)
			fname += r.URL.Path + ".html"
		} else {
			fname += r.URL.Path
		}
	}

	info, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			writeErrorHtml(r.Context(), w, http.StatusNotFound)
			return
		}

		writeErrorHtml(r.Context(), w, http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		writeErrorHtml(r.Context(), w, http.StatusNotFound)
		return
	}

	f, err := os.Open(fname)
	if err != nil {
		writeErrorHtml(r.Context(), w, http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}
