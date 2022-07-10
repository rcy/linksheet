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

const csvUrl = "https://docs.google.com/spreadsheets/d/e/2PACX-1vSdgZd_zTa2yympVBgqBTHWgyZ00_TtVV9BRYZDZKzNYo8ArtjqH6oTVlRbCWbbzl3Sg__f_kE9Pwg0/pub?gid=0&single=true&output=csv"

var csvmap map[string]string

var lastLookupAt time.Time

func Init(refresh time.Duration) error {
	log.Println("linkmap.Init", refresh)

	err := Sync()
	if err != nil {
		return err
	}

	go syncLoop(refresh)

	return nil
}

func syncLoop(refresh time.Duration) {
	for {
		time.Sleep(refresh)
		Sync()
	}	
}

func Sync() error {
	log.Println("linkmap.Sync")

	csvbytes, err := download(csvUrl)
	if err != nil {
		return err
	}
	csvmap, err = csv2map(csvbytes)
	if err != nil {
		return err
	}

	lastLookupAt = time.Now()

	return nil
}

func Lookup(alias string) (string) {
	target := csvmap[alias]
	log.Printf("linkmap.Lookup(%s) => %s", alias, target)
	return target
}

func download(url string) ([]byte, error) {
	client := http.Client{Timeout: time.Duration(10 * time.Second)}

	resp, err := client.Get(url)
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

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
