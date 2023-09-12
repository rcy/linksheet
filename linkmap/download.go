package linkmap

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func download(url string) ([]byte, error) {
	client := http.Client{Timeout: time.Duration(10 * time.Second)}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Status code: %d", resp.StatusCode))
	}

	contentType := resp.Header["Content-Type"][0]
	if contentType != "text/csv" {
		return nil, errors.New(fmt.Sprintf("Content-type: %s, want text/csv", contentType))
	}

	return b, nil
}
