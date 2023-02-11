package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"rcy/linksheet/db"
	"rcy/linksheet/linkmap"
	"strings"

	"github.com/gorilla/mux"
)

var Links *linkmap.LinkMap = nil

func main() {
	csvUrl, ok := os.LookupEnv("CSV_URL")
	if !ok {
		log.Fatalf("$CSV_URL not found in environment")
	}

	var err error
	Links, err = linkmap.Init(csvUrl)

	if err != nil {
		log.Fatalf("could not initialize linkmap from url %s: %s", csvUrl, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handleHome)
	r.HandleFunc("/_sync", handleSync)
	r.HandleFunc("/_requests", handleRequests)
	r.HandleFunc("/{alias}", handleLookup)

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

func readUserIP(r *http.Request) string {
	hostport := strings.Split(r.Header.Get("X-Forwarded-For"), ", ")[0]
	if hostport == "" {
		hostport = r.RemoteAddr
	}
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}
	return host
}

func handleLookup(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]
	target := Links.Lookup(alias)
	ip := readUserIP(r)

	log.Printf("%s|%s|%s", ip, alias, target)

	if target != "" {
		targetURL, err := url.Parse(target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// pass along the request query string, unless the target already has one
		if targetURL.RawQuery == "" {
			targetURL.RawQuery = r.URL.RawQuery
		}

		http.Redirect(w, r, targetURL.String(), http.StatusSeeOther)
		db.TrackRequest(ip, alias, target, http.StatusSeeOther)
	} else {
		str := fmt.Sprintf("%s not found\n", alias)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, str)
		db.TrackRequest(ip, alias, target, http.StatusNotFound)
	}
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	err := Links.Sync()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg := fmt.Sprintf("%d links", Links.Count())
	io.WriteString(w, msg)
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := db.Requests()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error")
		return
	}
	for _, r := range requests {
		io.WriteString(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", r.CreatedAt, r.Ip, r.Status, r.Alias, r.Target))
	}
}
