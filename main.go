package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"rcy/home/linkmap"

	"github.com/gorilla/mux"
)

var Links *linkmap.LinkMap = nil

func main() {
	csvUrl, ok := os.LookupEnv("GOOGLE_SHEET")
	if !ok {
		log.Fatalf("GOOGLE_SHEET not found in environment")
	}

	links, err := linkmap.Init(csvUrl)

	if err != nil {
		log.Fatalf("could not initialize linkmap from url %s: %s", csvUrl, err)
	}

	Links = links

	r := mux.NewRouter()
	r.HandleFunc("/", handleHome)
	r.HandleFunc("/_sync", handleSync)
	r.HandleFunc("/{alias}", handleAlias)

	http.Handle("/", r)

	fmt.Println("listening on port 3333")

	err = http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		os.Exit(1)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func handleAlias(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]
	target := Links.Lookup(alias)

	if target != "" {
		targetURL, err := url.Parse(target)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		// pass along the request query string, unless the target already has one
		if targetURL.RawQuery == "" {
			targetURL.RawQuery = r.URL.RawQuery
		}

		http.Redirect(w, r, targetURL.String(), http.StatusSeeOther)
	} else {
		str := fmt.Sprintf("%s not found\n", alias)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, str)
	}
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	err := Links.Sync()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	msg := fmt.Sprintf("%d links", Links.Count())
	io.WriteString(w, msg)
}
