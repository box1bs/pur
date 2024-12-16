package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func WriteJSON(w http.ResponseWriter, status int, mess any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(mess)
}

// Handle uuid from path
func varsHandleUUID(r *http.Request, key string) (uuid.UUID, error) {
	val := mux.Vars(r)[key]
	id, err := uuid.Parse(val)
	if err != nil {
		log.Printf("failed parsing sended id: %v", err)
		return uuid.Nil, err
	}
	return id, nil
}

type ApiError struct {
	Error string
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()}) // error handle
		}
	}
}

func CanSummarize(Type string) bool {
	return slices.Contains([]string{"article", "website", "book", "document", "event"}, Type) //constant types 
}