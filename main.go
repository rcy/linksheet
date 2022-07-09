package main

import (
	"net/http"
	"fmt"
	//	"io"
	"errors"
	"os"
)

func main() {
	http.HandleFunc("/zoom", getZoom)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
func getZoom(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("getZoom redirect\n")
	// io.WriteString(w, "This is my website!\n")
	http.Redirect(w, r, "https://www.gnu.org", http.StatusSeeOther)
}
