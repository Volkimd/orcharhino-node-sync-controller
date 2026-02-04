package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

type HostData struct {
	Name string `json:"name"`
}

type HostPayload struct {
	Host HostData `json:"host"`
}

// In-Memory Speicher für unsere Nodes
var (
	inventory = make(map[string]bool)
	mu        sync.Mutex
)

func main() {
	http.HandleFunc("/api/hosts", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == "POST" {
			var payload HostPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "Bad Request", 400)
				return
			}
			inventory[payload.Host.Name] = true
			fmt.Printf("✅ ADD: Node '%s' registriert. Aktueller Bestand: %v\n", payload.Host.Name, getInventoryKeys())
			w.WriteHeader(http.StatusCreated)

		} else if r.Method == "GET" {
			json.NewEncoder(w).Encode(getInventoryKeys())
		}
	})

	// DELETE Endpunkt: /api/hosts/{name}
	http.HandleFunc("/api/hosts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			mu.Lock()
			defer mu.Unlock()

			nodeName := strings.TrimPrefix(r.URL.Path, "/api/hosts/")
			if _, exists := inventory[nodeName]; exists {
				delete(inventory, nodeName)
				fmt.Printf("❌ DELETE: Node '%s' entfernt. Aktueller Bestand: %v\n", nodeName, getInventoryKeys())
				w.WriteHeader(http.StatusOK)
			} else {
				fmt.Printf("⚠️ DELETE: Node '%s' nicht gefunden.\n", nodeName)
				w.WriteHeader(http.StatusNotFound)
			}
		}
	})

	fmt.Println("🚀 orcharhino Mock-Server läuft auf http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getInventoryKeys() []string {
	keys := make([]string, 0, len(inventory))
	for k := range inventory {
		keys = append(keys, k)
	}
	return keys
}
