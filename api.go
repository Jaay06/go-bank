package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)



type APIServer struct {
	listenAddr string
}

//api server
func NewAPIServer(listenAddr string) *APIServer{
	return &APIServer{
		listenAddr: listenAddr,
	}
}

///server run function
func (s *APIServer) Run(){
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", makeHTTPHandleFunc(s.handleGetAccount))


	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

//handle all account function "GET", "POST", "DELETE"
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error{

	switch r.Method {
	case "GET": return s.handleGetAccount(w, r)

	case "POST": return s.handleCreateAccount(w, r)
	
	case "DELETE": return s.handleDeleteAccount(w, r)

	default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return nil
	}
	
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error{
	id := mux.Vars(r)["id"]
	
	// account := NewAccount("Jaay", "Tee")
	fmt.Println(id)


	return WriteJSON(w, http.StatusOK, &Account{})
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error{
	return nil
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error{
	return nil
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error{
	return nil
}


/*helper functions */
func WriteJSON(w http.ResponseWriter, status int, v any) error{
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		if err := f(w, r); err != nil{
			//handle error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			
		}
	}
}