package parser

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	lock  sync.RWMutex
	proto Protocols
)

type Protocol struct {
	Name string
	Tags []string
	ASN  int
}

type Protocols map[string]Protocol

// LoadProtocols updates the protocols map from a JSON file
func LoadProtocols(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var protocols Protocols
	if err := json.Unmarshal(b, &protocols); err != nil {
		return err
	}

	lock.Lock()
	defer lock.Unlock()
	proto = protocols

	return nil
}

// GetProtocol gets a protocol by name
func GetProtocol(name string) Protocol {
	lock.RLock()
	defer lock.RUnlock()
	p, ok := proto[name]
	if !ok {
		p = Protocol{
			Name: name,
			Tags: []string{},
			ASN:  0,
		}
	}

	if strings.HasSuffix(name, "_v4") || strings.Contains(name, "_v4_") {
		p.Name += " IPv4"
	} else if strings.HasSuffix(name, "_v6") || strings.Contains(name, "_v6_") {
		p.Name += " IPv6"
	}

	// Parse ASN from protocol name
	nameParts := strings.Split(name, "_")
	var asnStr string
	for _, part := range nameParts {
		if strings.HasPrefix(part, "AS") {
			asnStr = part[2:] // trim AS prefix
			break
		}
	}
	if asnStr != "" {
		p.ASN = int(parseInt(asnStr))
	}

	return p
}

// WatchProtocols watches a file for changes and updates the protocols map
func WatchProtocols(filename string, watcher *fsnotify.Watcher) error {
	if err := LoadProtocols(filename); err != nil {
		log.Printf("Failed to update protocols: %s", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					if err := LoadProtocols(filename); err != nil {
						log.Printf("Failed to update protocols: %s", err)
					}
				}
			}
		}
	}()

	return watcher.Add(filename)
}
