package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/csv"
	"hash"
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
	if parsed[2] == base64.StdEncoding.EncodeToString(digest.Sum(nil)) {
		return true
	}

	return false
}
