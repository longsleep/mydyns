package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

type nsUpdateData struct {
	hostname string
	ip       *net.IP
}

type NsUpdate struct {
	exe     string
	server  string
	keyfile string
	zone    string
	ttl     int
	queue   chan *nsUpdateData
	exit    chan bool
	timer   chan bool
}

func NewNsUpdate(exe, server, keyfile, zone string, ttl int) *NsUpdate {
	return &NsUpdate{
		exe:     exe,
		server:  server,
		keyfile: keyfile,
		zone:    zone,
		ttl:     ttl,
		queue:   make(chan *nsUpdateData, 100),
		exit:    make(chan bool),
	}
}

func (update *NsUpdate) run() {
	work := make(map[string]*net.IP)
	var err error
	c := time.Tick(5 * time.Second)
	for {
		select {
		case <-c:
		Work:
			for {
				select {
				case data := <-update.queue:
					log.Println("Processing update", data.hostname, data.ip)
					work[data.hostname] = data.ip
				default:
					// No data available. Non blocking.
					break Work
				}
			}
			if len(work) > 0 {
				// Do some work.
				err = update.process(work)
				if err != nil {
					// Error.
					log.Println("Update failed", err)
				} else {
					work = make(map[string]*net.IP)
				}
			}
		case <-update.exit:
			return
		}
	}
}

func (update *NsUpdate) process(work map[string]*net.IP) error {

	f, err := ioutil.TempFile(os.TempDir(), "mydyndnsrv")
	if err != nil {
		return err
	}
	log.Printf("Processing %d updates in %s", len(work), f.Name())
	defer os.Remove(f.Name())
	w := bufio.NewWriter(f)

	w.WriteString(fmt.Sprintf("server %s\n", update.server))
	w.WriteString(fmt.Sprintf("zone %s\n", update.zone))

	var recordtype string
	for hostname, ip := range work {
		if ip.To4() != nil {
			recordtype = "A"
		} else {
			recordtype = "AAAA"
		}
		w.WriteString(fmt.Sprintf("update delete %s.%s. %s\n", hostname, update.zone, recordtype))
		w.WriteString(fmt.Sprintf("update add %s.%s. %d %s %s\n", hostname, update.zone, update.ttl, recordtype, ip))
	}

	w.WriteString("send\n")

	w.Flush()
	f.Close()

	// Run command.
	cmd := exec.Command(update.exe, "-k", update.keyfile, f.Name())
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return err
	}
	log.Println("Completed update", f.Name())
	return nil

}

func (update *NsUpdate) update(data *nsUpdateData) error {
	// Send non blocking.
	select {
	case update.queue <- data:
		return nil
	default:
		return errors.New("update queue full")
	}
}
