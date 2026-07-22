package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type metric struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Change string `json:"change"`
	Trend  string `json:"trend"`
}

type project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Status     string `json:"status"`
	Completion int    `json:"completion"`
}

type activity struct {
	ID       string `json:"id"`
	Initials string `json:"initials"`
	Person   string `json:"person"`
	Action   string `json:"action"`
	When     string `json:"when"`
}

type dashboard struct {
	Metrics     []metric   `json:"metrics"`
	Projects    []project  `json:"projects"`
	Activity    []activity `json:"activity"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

type server struct {
	mu             sync.RWMutex
	projects       []project
	activity       []activity
	nextProjectID  int
	nextActivityID int
}

func main() {
	s := &server{
		projects: []project{
			{ID: "project-1", Name: "Customer portal", Owner: "Maya Chen", Status: "On track", Completion: 82},
			{ID: "project-2", Name: "Mobile analytics", Owner: "Theo Martin", Status: "At risk", Completion: 46},
			{ID: "project-3", Name: "Q3 launch", Owner: "Ava Robinson", Status: "On track", Completion: 68},
			{ID: "project-4", Name: "Design system", Owner: "Noah Williams", Status: "Review", Completion: 91},
		},
		activity: []activity{
			{ID: "activity-1", Initials: "MC", Person: "Maya Chen", Action: "shipped the new onboarding flow", When: "12 min ago"},
			{ID: "activity-2", Initials: "TM", Person: "Theo Martin", Action: "flagged a mobile release blocker", When: "38 min ago"},
			{ID: "activity-3", Initials: "AR", Person: "Ava Robinson", Action: "updated the Q3 launch brief", When: "1 hr ago"},
		},
		nextProjectID:  5,
		nextActivityID: 4,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/dashboard", s.getDashboard)
	mux.HandleFunc("POST /api/projects", s.createProject)
	mux.HandleFunc("PATCH /api/preferences", s.updatePreferences)

	log.Println("dashboard API listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", cors(mux)))
}

func (s *server) getDashboard(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	respond(w, http.StatusOK, dashboard{
		Metrics: []metric{
			{Label: "Total revenue", Value: "$48,240", Change: "+12.5%", Trend: "up"},
			{Label: "Active users", Value: "2,350", Change: "+8.2%", Trend: "up"},
			{Label: "Conversion rate", Value: "3.24%", Change: "+0.4%", Trend: "up"},
			{Label: "Open tickets", Value: "18", Change: "-14.3%", Trend: "down"},
		},
		Projects:    s.projects,
		Activity:    s.activity,
		GeneratedAt: time.Now().UTC(),
	})
}

func (s *server) createProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		http.Error(w, "a project name is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	created := project{
		ID:         fmt.Sprintf("project-%d", s.nextProjectID),
		Name:       input.Name,
		Owner:      "You",
		Status:     "Planning",
		Completion: 5,
	}
	s.nextProjectID++
	createdActivity := activity{
		ID:       fmt.Sprintf("activity-%d", s.nextActivityID),
		Initials: "YO",
		Person:   "You",
		Action:   "created " + input.Name,
		When:     "just now",
	}
	s.nextActivityID++
	s.projects = append([]project{created}, s.projects...)
	s.activity = append([]activity{createdActivity}, s.activity...)
	respond(w, http.StatusCreated, created)
}

func (s *server) updatePreferences(w http.ResponseWriter, r *http.Request) {
	var preference struct {
		DailyDigest bool `json:"dailyDigest"`
	}
	if err := json.NewDecoder(r.Body).Decode(&preference); err != nil {
		http.Error(w, "invalid preference", http.StatusBadRequest)
		return
	}
	respond(w, http.StatusOK, preference)
}

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
