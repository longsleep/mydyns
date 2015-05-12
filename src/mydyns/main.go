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
	"fmt"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var version = "0.0.1"
var (
	_, IPv4privateA, _ = net.ParseCIDR("10.0.0.0/8")
	_, IPv4privateB, _ = net.ParseCIDR("172.16.0.0/12")
	_, IPv4privateC, _ = net.ParseCIDR("192.168.0.0/16")
	_, IPv6private, _  = net.ParseCIDR("fd00::/8")
)

var update *NsUpdate
var users *HtpasswdFile
var hosts *HostsFile
var secret *SecretFile

// TokenData defines the data to encode into tokens.
type TokenData struct {
	Host string
	User string
}

// isPrivateNetwork checks if an IP address is inside a private network.
func isPrivateNetwork(ip net.IP) bool {
	if IPv4privateA.Contains(ip) ||
		IPv4privateB.Contains(ip) ||
		IPv4privateC.Contains(ip) ||
		IPv6private.Contains(ip) {
		return true
	}
	return false
}

// main is our blocking runner.
func main() {

	// Parse command line.
	var (
		listen     = kingpin.Flag("listen", "Listen address.").PlaceHolder("IP:PORT").Default("127.0.0.1:8080").String()
		nsupdate   = kingpin.Flag("nsupdate", "Path to nsupdate binary.").Default("/usr/bin/nsupdate").ExistingFile()
		server     = kingpin.Flag("server", "DNS server hostname.").Required().String()
		keyfile    = kingpin.Flag("key", "DNS shared secrets file.").Required().PlaceHolder("KEYFILE").ExistingFile()
		zone       = kingpin.Flag("zone", "Zone where updates should be made.").Required().String()
		ttl        = kingpin.Flag("ttl", "Ttl for DNS entries.").Default("300").Int()
		usersfile  = kingpin.Flag("users", "Htpasswd users database.").Required().PlaceHolder("USERSFILE").ExistingFile()
		hostsfile  = kingpin.Flag("hosts", "Hosts database.").Required().PlaceHolder("HOSTSFILE").ExistingFile()
		secretfile = kingpin.Flag("secret", "Auth token secret file.").Required().ExistingFile()
	)

	kingpin.CommandLine.Help = "Manage your own dynamic DNS zone."
	kingpin.Version(version)
	kingpin.Parse()

	log.Printf("Starting up on: %s\n", *listen)

	// Initialize.
	update = NewNsUpdate(*nsupdate, *server, *keyfile, *zone, *ttl)
	users, _ = NewHtpasswdFile(*usersfile)
	hosts, _ = NewHostsFile(*hostsfile)
	secret, _ = NewSecretFile(*secretfile)

	// Create URL routing.
	mux := http.NewServeMux()
	mux.HandleFunc("/update", updateHandler)
	mux.HandleFunc("/token", tokenHandler)

	// Start our worker.
	go update.run()

	// Start HTTP service.
	s := &http.Server{
		Addr:           *listen,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())

}

// updateHandler implementes the end point to update IP address for a given token.
func updateHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	token := r.Form.Get("token")
	myip := r.Form.Get("myip")

	// Validate token.
	if token == "" {
		http.Error(w, "token parameter required", http.StatusBadRequest)
		return
	}
	var data TokenData
	if err := secret.Decode("u", token, &data); err != nil {
		http.Error(w, fmt.Sprintf("invalid token: %s", err), http.StatusForbidden)
		return
	}

	// Validate hostname access in users database.
	if !hosts.CheckUser(data.Host, data.User) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// Get IP.
	var ip net.IP
	if myip == "" || myip == "auto" {
		myip = strings.SplitN(r.RemoteAddr, ":", 2)[0]
		ip = net.ParseIP(myip)
		if ip.IsLoopback() {
			// Running through a proxy?
			myip = r.Header.Get("X-Real-IP")
			if myip != "" {
				ip = net.ParseIP(myip)
			}
		}
	} else {
		ip = net.ParseIP(myip)
	}
	// Validate IP.
	if ip == nil || !ip.IsGlobalUnicast() {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	} else if isPrivateNetwork(ip) {
		http.Error(w, "private ip not allowed", http.StatusBadRequest)
		return
	}

	if _, ok := r.Form["check"]; ok {
		fmt.Fprintf(w, fmt.Sprintf("%s\n", ip))
		return
	}

	// Queue changes.
	if err := update.update(&nsUpdateData{data.Host, &ip}); err != nil {
		log.Println("Update failed", err)
		http.Error(w, fmt.Sprintf("update failed: %s", err), http.StatusTeapot)
	} else {
		log.Println("Queued update", data.Host, ip)
	}

	fmt.Fprintf(w, "accepted\n")

}

// tokenHandler creates tokens for a given hostname.
func tokenHandler(w http.ResponseWriter, r *http.Request) {

	// Basic auth is required.
	username, password, ok := getBasicAuth(r)
	if ok {
		if !users.CheckPassword(username, password) {
			http.Error(w, "authentication failed", http.StatusForbidden)
			return
		}
	} else {
		http.Error(w, "basic auth required", http.StatusForbidden)
		return
	}

	r.ParseForm()
	hostname := r.Form.Get("hostname")
	if hostname == "" {
		http.Error(w, "hostname parameter required", http.StatusBadRequest)
		return
	}

	// Validate hostname.
	if url, err := url.Parse(fmt.Sprintf("http://%s/", hostname)); err != nil {
		http.Error(w, fmt.Sprintf("invalid hostname: %s", err), http.StatusBadRequest)
		return
	} else {
		host := strings.SplitN(url.Host, ":", 2)[0]
		host = strings.SplitN(host, ".", 2)[0]
		if host != hostname {
			http.Error(w, "invalid hostname", http.StatusBadRequest)
			return
		}
	}

	// Validate hostname access in users database.
	if !hosts.CheckUser(hostname, username) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// Prepare and encode token.
	data := &TokenData{
		Host: hostname,
		User: username,
	}
	if token, err := secret.Encode("u", data); err == nil {
		log.Println("Token created by", username, hostname)
		fmt.Fprintln(w, token)
	} else {
		log.Println("Error while creating token", err)
		http.Error(w, fmt.Sprintf("failed to create token: %s", err), http.StatusInternalServerError)
	}

}