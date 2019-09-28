/*
Mydyns - run your own dynamic DNS zone
Copyright (C) 2015  Simon Eisenmann

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

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
