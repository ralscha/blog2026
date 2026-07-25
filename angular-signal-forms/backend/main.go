package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var takenUsernames = map[string]bool{
	"admin":       true,
	"root":        true,
	"fonduepilot": true,
	"matterhorn":  true,
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/usernames/", usernameAvailability)

	log.Println("Signal Forms demo backend listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", withCors(mux)); err != nil {
		log.Fatal(err)
	}
}

func usernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/usernames/")
	username = strings.ToLower(strings.TrimSpace(username))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{
		"available": username != "" && !takenUsernames[username],
	})
}

func withCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
