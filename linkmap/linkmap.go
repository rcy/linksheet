package linkmap

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

type LinkMap struct {
	url      string
	csvbytes []byte
	csvmap   map[string]string
}

func NewFromCSVString(csv string) *LinkMap {
	m := &LinkMap{csvbytes: []byte(csv)}
	m.csv2map()
	return m
}

func NewFromURL(url string) (*LinkMap, error) {
	m := &LinkMap{url: url}

	err := m.Sync()
	if err != nil {
		return m, err
	}

	go m.loop(600 * time.Second)

	return m, nil
}

var re = regexp.MustCompile("(.+)([*])")

func (m *LinkMap) Lookup(input string) string {
	res := m.csvmap[input]
	if res != "" {
		return res
	}
	for pat, target := range m.csvmap {
		fmt.Printf("%s %s %s\n", input, pat, target)
		// pat    = wild/*
		// target = example.com/*
		// input  = wild/12345
		// output = example.com/12345

		patMatches := re.FindStringSubmatch(pat) // wild/, *
		if patMatches == nil {
			fmt.Printf("patMatches: %s\n", patMatches)
			continue
		}
		targetMatches := re.FindStringSubmatch(target) // example.com/, *
		if targetMatches == nil {
			fmt.Printf("targetMatches: %s\n", targetMatches)
			continue
		}
		replacement := strings.TrimPrefix(input, patMatches[1]) // wild/12345 -> 12345
		if replacement == input {
			continue
		}

		return targetMatches[1] + replacement // example.com/12345
	}
	return ""
}

func (m *LinkMap) loop(refresh time.Duration) {
	for {
		time.Sleep(refresh)
		m.Sync()
	}
}

func (m *LinkMap) Sync() error {
	log.Println("linkmap.Sync")

	bytes, err := download(m.url)
	if err != nil {
		return err
	}
	m.csvbytes = bytes

	err = m.csv2map()
	if err != nil {
		return err
	}

	log.Printf("synced %v items", len(m.csvmap))

	return nil
}

func (m *LinkMap) Count() int {
	return len(m.csvmap)
}

func (m *LinkMap) csv2map() error {
	res := map[string]string{}
	reader := bytes.NewReader(m.csvbytes)
	data, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return err
	}
	for _, row := range data {
		res[row[0]] = row[1]
	}
	m.csvmap = res
	return nil
}
