package linkmap

import (
	"bytes"
	"encoding/csv"
	"log"
	"time"
)

type LinkMap struct {
	Url      string
	csvbytes []byte
	csvmap   map[string]string
}

func NewFromURL(url string) (*LinkMap, error) {
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

	bytes, err := download(m.Url)
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
