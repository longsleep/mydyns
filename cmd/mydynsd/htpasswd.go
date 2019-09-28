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
	"crypto/sha1"
	"crypto/subtle"
	"encoding/csv"
	"hash"
	"log"
	"os"
	"regexp"
)

// passwordParse defines a regular expression to get the password and hash.
var passwordParser, _ = regexp.Compile("{([A-Z]+)}(.*)")

type HtpasswdFile struct {
	users map[string]string
}

func NewHtpasswdFile(fn string) (*HtpasswdFile, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// htpasswd files are essentially csv files.
	reader := csv.NewReader(f)
	reader.Comma = ':'
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	entries, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	ht := &HtpasswdFile{
		users: make(map[string]string),
	}
	for _, entry := range entries {
		ht.users[entry[0]] = entry[1]
	}

	log.Printf("Loaded %d users\n", len(ht.users))
	return ht, nil
}

func (ht *HtpasswdFile) CheckPassword(user, password string) bool {
	entry, ok := ht.users[user]
	if !ok {
		return false
	}

	// Parse password entry into hash type and value.
	parsed := passwordParser.FindStringSubmatch(entry)
	if len(parsed) < 3 {
		return false
	}

	// Switch by hash.
	var digest hash.Hash
	switch parsed[1] {
	case "SHA":
		digest = sha1.New()
	default:
		return false
	}

	digest.Write([]byte(password))
	return subtle.ConstantTimeCompare([]byte(parsed[2]), digest.Sum(nil)) == 1
}
