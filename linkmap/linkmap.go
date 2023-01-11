package linkmap

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type LinkMap struct {
	Url          string
	csvmap       map[string]string
	lastLookupAt time.Time
}

func Init(url string) (*LinkMap, error) {
	m := &LinkMap{Url: url}

	err := m.Sync()
	if err != nil {
		return m, err
	}

	go m.loop(600 * time.Second)

	return m, nil
}

func (m *LinkMap) Lookup(alias string) string {
	return m.csvmap[alias]
}

func (m *LinkMap) loop(refresh time.Duration) {
	for {
		time.Sleep(refresh)
		m.Sync()
	}
}

func (m *LinkMap) Sync() error {
	log.Println("linkmap.Sync")

	csvbytes, err := download(m.Url)
	if err != nil {
		return err
	}
	m.csvmap, err = csv2map(csvbytes)
	if err != nil {
		return err
	}

	m.lastLookupAt = time.Now()

	log.Printf("synced %v items", len(m.csvmap))

	return nil
}

func (m *LinkMap) Count() int {
	return len(m.csvmap)
}

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

func csv2map(input []byte) (map[string]string, error) {
	res := map[string]string{}
	reader := bytes.NewReader(input)
	data, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return nil, err
	}
	for _, row := range data[1:] {
		res[row[0]] = row[1]
	}
	return res, nil
}
