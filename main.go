package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"rcy/home/linkmap"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	linkmap.Init(600 * time.Second)

	r := mux.NewRouter()
	r.HandleFunc("/_sync", handleSync)
	r.HandleFunc("/{alias}", handleAlias)

	http.Handle("/", r)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		os.Exit(1)
	}
}

func handleAlias(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]
	target := linkmap.Lookup(alias)

	if target != "" {
		targetURL, err := url.Parse(target)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		// pass along the original query
		targetURL.RawQuery = r.URL.RawQuery

		http.Redirect(w, r, targetURL.String(), http.StatusSeeOther)
	} else {
		str := fmt.Sprintf("%s not found\n", alias)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, str)
	}
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	err := linkmap.Sync()
	if err != nil {
		w.WriteHeader(500)
	}

	io.WriteString(w, "sync")
}
