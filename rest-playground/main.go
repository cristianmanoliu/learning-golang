package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// User represents a simple user entity.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// userStore is an in-memory store with basic concurrency protection.
type userStore struct {
	// sync.Mutex is Go’s simplest mutual exclusion lock: only one goroutine can hold it at a time.
	mu     sync.Mutex
	users  []User
	nextID int
}

func newUserStore() *userStore {
	return &userStore{
		users:  make([]User, 0),
		nextID: 1,
	}
}

// addUser inserts a new user with a generated ID.
func (s *userStore) addUser(name string) User {
	// Lock the mutex to ensure exclusive access to the users slice.
	s.mu.Lock()
	// Release the lock when the function returns.
	defer s.mu.Unlock()

	u := User{
		ID:   s.nextID,
		Name: name,
	}
	s.nextID++
	s.users = append(s.users, u)
	return u
}

// listUsers returns a copy of all users.
func (s *userStore) listUsers() []User {
	// func (s userStore) ... copies sync.Mutex → lock/unlock the wrong thing → racy + broken.
	// func (s *userStore) ... copies sync.Mutex → lock/unlock the wrong thing → racy + broken.
	// Once you have a mutex in a struct, always use pointer receivers for methods that touch it.
	// Lock the mutex to ensure exclusive access to the users slice.
	s.mu.Lock()
	// Defer the unlock to ensure it happens when the function returns.
	defer s.mu.Unlock()

	// Return a copy to prevent external modification.
	result := make([]User, len(s.users))
	copy(result, s.users)
	return result
}

func main() {
	log.Println("starting REST playground on :8080")

	store := newUserStore()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateUser(w, r, store)
		case http.MethodGet:
			handleListUsers(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Optional: individual user by id (GET /users/{id})
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/users/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		users := store.listUsers()
		for _, u := range users {
			if u.ID == id {
				respondJSON(w, http.StatusOK, u)
				return
			}
		}

		http.Error(w, "not found", http.StatusNotFound)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

type createUserRequest struct {
	Name string `json:"name"`
}

func handleCreateUser(w http.ResponseWriter, r *http.Request, store *userStore) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	u := store.addUser(req.Name)
	respondJSON(w, http.StatusCreated, u)
}

func handleListUsers(w http.ResponseWriter, _ *http.Request, store *userStore) {
	users := store.listUsers()
	respondJSON(w, http.StatusOK, users)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode json: %v", err)
	}
}
