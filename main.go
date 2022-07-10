package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"rcy/home/linkmap"
	"time"
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
		http.Redirect(w, r, target, http.StatusSeeOther)
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
