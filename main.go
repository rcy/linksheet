package main

import (
	"net/http"
	"fmt"
	//	"io"
	"errors"
	"os"
)

func main() {
	http.HandleFunc("/sesh", getSesh)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
func getSesh(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://us02web.zoom.us/j/3499596140?pwd=bldEUStXYWFKM3pUR3R0TlhwdE9tQT09", http.StatusSeeOther)
}
