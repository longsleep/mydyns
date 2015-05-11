package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

type HostsFile struct {
	hosts map[string][]string
}

func NewHostsFile(fn string) (*HostsFile, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// hosts files are essentially csv files.
	reader := csv.NewReader(f)
	reader.Comma = ':'
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	entries, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	h := &HostsFile{
		hosts: make(map[string][]string),
	}
	for _, entry := range entries {
		h.hosts[entry[0]] = strings.Split(entry[1], ",")
	}

	log.Printf("Loaded %d hosts\n", len(h.hosts))
	return h, nil
}

func (h *HostsFile) CheckUser(host, user string) bool {
	entry, ok := h.hosts[host]
	if !ok {
		return false
	}
	for _, u := range entry {
		if u == user {
			return true
		}
	}
	return false
}
