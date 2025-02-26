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

	"github.com/go-chi/chi/v5"
	"github.com/vincent-petithory/dataurl"
)

var Links *linkmap.LinkMap = nil
var Password string

func main() {
	var ok bool
	Password, ok = os.LookupEnv("PASSWORD")
	if !ok {
		log.Fatalf("$PASSWORD not found in environment")
	}

	csvUrl, ok := os.LookupEnv("CSV_URL")
	if !ok {
		log.Fatalf("$CSV_URL not found in environment")
	}

	var err error
	Links, err = linkmap.NewFromURL(csvUrl)

	if err != nil {
		log.Fatalf("could not initialize linkmap from url %s: %s", csvUrl, err)
	}

	r := chi.NewRouter()
	r.Get("/", handleHome)
	r.Get("/favicon.ico", handleFavIcon)
	r.Get("/_sync", handleSync)
	r.Get("/_requests", handleRequests)
	r.Get("/linksheet.db", withAuth(handleDb))
	r.Get("/*", handleLookup)

	http.Handle("/", r)

	log.Print("listening on port 3333")

	err = http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	} else if err != nil {
		log.Printf("server closed unexpectedly: %v\n", err)
		os.Exit(1)
	}
}

func withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "admin" || p != Password {
			w.Header().Add("WWW-Authenticate", `Basic realm="hold up"`)
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "unauthorized")
			return
		}
		handler(w, r)
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

func handleFavIcon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func handleLookup(w http.ResponseWriter, r *http.Request) {
	alias := strings.TrimPrefix(r.URL.Path, "/")
	target := Links.Lookup(alias)
	ip := readUserIP(r)

	log.Printf("%s|%s|%s", ip, alias, target)

	if target != "" {
		targetURL, err := url.Parse(target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch targetURL.Scheme {
		case "data":
			// serve the content of data urls directly
			dataURL, err := dataurl.DecodeString(targetURL.String())
			if err != nil {
				http.Error(w, "DecodeString: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Add("Content-Type", dataURL.ContentType())
			_, err = w.Write(dataURL.Data)
			if err != nil {
				http.Error(w, "Write: "+err.Error(), http.StatusInternalServerError)
			}
			db.TrackRequest(ip, alias, target, http.StatusOK)
		default:
			// pass along the request query string, unless the target already has one
			if targetURL.RawQuery == "" {
				targetURL.RawQuery = r.URL.RawQuery
			}

			http.Redirect(w, r, targetURL.String(), http.StatusSeeOther)
		}
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

func handleDb(w http.ResponseWriter, r *http.Request) {
	filename, ok := os.LookupEnv("DB_FILE")
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal server error")
		return
	}

	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	size := stat.Size()

	// Read the contents of the file
	buffer := make([]byte, size)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the contents of the file to the ResponseWriter
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=linksheet.db")
	w.Write(buffer)
}
