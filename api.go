package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)



type APIServer struct {
	listenAddr string
	store Storage
}

//api server
func NewAPIServer(listenAddr string, store Storage) *APIServer{
	return &APIServer{
		listenAddr: listenAddr,
		store: store,
	}
}

///server run function
func (s *APIServer) Run(){
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store))

	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))


	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

//handle all account function "GET", "POST", "DELETE"
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error{

	switch r.Method {
	case "GET": return s.handleGetAccount(w, r)

	case "POST": return s.handleCreateAccount(w, r)
	

	default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return nil
	}
	
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error{

	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err :=json.NewDecoder(r.Body).Decode(&req); err != nil {
	return err
	}

	acc, err := s.store.GetAccountByNumber(int(req.Number))

	if err != nil {
		return err //handle this response as json
	}

	if !acc.ValidatePassword(req.Password){
		return fmt.Errorf("not authenticated")
	}

	//create JWT
	token, err := createJWT(acc)
	if err != nil {
		return err
	}

	response := LoginResponse{
		Token: token,
		Number: acc.Number,
	}



return WriteJSON(w, http.StatusOK, response)
	
}

/**get all account on the database */
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error{
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error{
	//get by id functionality
	if r.Method == "GET"{
		//converts id to a number
		id, err := getID(r)
		if err != nil {
			return err
		}
		
		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}
		
		return WriteJSON(w, http.StatusOK, account)
	} else if r.Method == "DELETE" { 
	//delete by id functionality
		return s.handleDeleteAccount(w, r)
	}




	return fmt.Errorf("method not allowed %s", r.Method)

}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error{
	req := new(CreateAccountRequest)
	// createAccountReq := CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	};

	account, err := NewAccount(req.FirstName, req.LastName, req.Password)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error{

	id, err := getID(r)
	if err != nil {
		return err
	}

	deleteErr := s.store.DeleteAccount(id)

if deleteErr != nil {
	return deleteErr
}

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error{
	transferReq := new (TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferReq)
}


/*helper functions */
func WriteJSON(w http.ResponseWriter, status int, v any) error{
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func createJWT (account *Account) (string, error){

	secret := os.Getenv("JWT_SECRET")

	// Create the Claims
	claims := &jwt.MapClaims{
	"expiresAt": 		jwt.NewNumericDate(time.Unix(1516239022, 0)),
	"accountNumber": 	account.Number,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func permissionDeniend(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})

}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request){

	tokenString := r.Header.Get("x-jwt-token")

	token, err :=validateJWT(tokenString)

	if err != nil {
		permissionDeniend(w)
		return
	}

	if !token.Valid {
		permissionDeniend(w)
		return
	}

	userID, err := getID(r)
	if err != nil {
		permissionDeniend(w)
		return
	}

	
	account, err := s.GetAccountByID(userID)
	if err != nil {
		permissionDeniend(w)
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	if account.Number != int64(claims["accountNumber"].(float64)){ //converts the clamims account number from float^4 to int 64
		permissionDeniend(w)
		return
	}


		handlerFunc(w, r)
	}
}



func validateJWT(tokenString string) (*jwt.Token, error) {

	jwtSecret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	

}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		if err := f(w, r); err != nil{
			//handle error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			
		}
	}
}

func getID(r *http.Request) (int, error) {

	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	
	return id, nil

}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo4MTM2NTI3NCwiZXhwaXJlc0F0IjoxNTE2MjM5MDIyfQ.Uerenr4PNA_KgmeEuZh6Ko_esuiYN-Ugh3OI1Hk7TOM