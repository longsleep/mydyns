package main

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var tokens *securecookie.SecureCookie
var (
	_, IPv4privateA, _ = net.ParseCIDR("10.0.0.0/8")
	_, IPv4privateB, _ = net.ParseCIDR("172.16.0.0/12")
	_, IPv4privateC, _ = net.ParseCIDR("192.168.0.0/16")
	_, IPv6private, _  = net.ParseCIDR("fd00::/8")
)

var update *nsUpdate

// TokenData defines the data to encode into tokens.
type TokenData struct {
	Hostname string
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

	// Initialize.
	tokens = securecookie.New([]byte("very-secret"), nil)
	update = newNsUpdate("path/to/nsupdate", "ns1.example.com", "/path/to/key.asc", "dyn.example.com", 300)

	// Create URL routing.
	mux := http.NewServeMux()
	mux.HandleFunc("/update", updateHandler)
	mux.HandleFunc("/token", tokenHandler)

	// Start our worker.
	go update.run()

	// Start HTTP service.
	http.ListenAndServe(":8081", mux)

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
	if err := tokens.Decode("u", token, &data); err != nil {
		http.Error(w, fmt.Sprintf("invalid token: %s", err), http.StatusForbidden)
		return
	}

	// Get IP.
	if myip == "" || myip == "auto" {
		myip = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}
	// Validate IP.
	ip := net.ParseIP(myip)
	if ip == nil || !ip.IsGlobalUnicast() {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	} else if isPrivateNetwork(ip) {
		http.Error(w, "private ip not allowed", http.StatusBadRequest)
		return
	}

	// Queue changes.
	if err := update.update(&nsUpdateData{data.Hostname, &ip}); err != nil {
		log.Println("Update failed", err)
		http.Error(w, fmt.Sprintf("update failed: %s", err), http.StatusTeapot)
	} else {
		log.Println("Queued update", data.Hostname, ip)
	}

	fmt.Fprintf(w, "accepted\n")

}

// tokenHandler creates tokens for a given hostname.
func tokenHandler(w http.ResponseWriter, r *http.Request) {

	// TODO(longsleep): Add basic auth support.

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

	// Prepare and encode token.
	data := &TokenData{
		Hostname: hostname,
	}
	if token, err := tokens.Encode("u", data); err == nil {
		log.Println("Token created", hostname)
		fmt.Fprintln(w, token)
	} else {
		log.Println("Error while creating token", err)
		http.Error(w, fmt.Sprintf("failed to create token: %s", err), http.StatusInternalServerError)
	}

}
