package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

var (
	getStudent      = regexp.MustCompile(`^\/users\/(\d+)$`)
	RegisterStudent = regexp.MustCompile(`^\/users[\/]*$`)
)

type Student struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type datastore struct {
	m map[string]Student
	*sync.RWMutex
}

type StudentHandler struct {
	store *datastore
}

func (h *StudentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet && getStudent.MatchString(r.URL.Path):
		h.Get(w, r)
		return
	case r.Method == http.MethodPost && RegisterStudent.MatchString(r.URL.Path):
		h.Create(w, r)
		return
	default:
		notFound(w, r)
		return
	}
}

func (h *StudentHandler) List(w http.ResponseWriter, r *http.Request) {
	h.store.RLock()
	students := make([]Student, 0, len(h.store.m))
	for _, v := range h.store.m {
		students = append(students, v)
	}
	h.store.RUnlock()
	jsonBytes, err := json.Marshal(students)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *StudentHandler) Get(w http.ResponseWriter, r *http.Request) {
	matches := getStudent.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	h.store.RLock()
	u, ok := h.store.m[matches[1]]
	h.store.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var student Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		internalServerError(w, r)
		return
	}
	h.store.Lock()
	str := strconv.Itoa(student.ID)
	h.store.m[str] = student
	h.store.Unlock()
	jsonBytes, err := json.Marshal(student)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal server error"))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func main() {
	mux := http.NewServeMux()
	userH := &StudentHandler{
		store: &datastore{
			m: map[string]Student{
				"Data": Student{
					ID:   1,
					Name: "bob",
					Age:  21,
				},
			},
			RWMutex: &sync.RWMutex{},
		},
	}
	mux.Handle("/Student", userH)
	mux.Handle("/Student/{id}", userH)

	http.ListenAndServe("localhost:8080", mux)
}
