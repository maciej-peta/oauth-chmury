// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
)

// User represents a row in the "users" table.
type User struct {
	AuthID        string `json:"auth_id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	AccountTypeID string `json:"account_type_id"`
}

var db *sql.DB

func createOrUpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	setHandlerHeaders(w, r, "POST", "OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only POST is allowed on this endpoint.")
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Upsert query (assumes a UNIQUE constraint on users(auth_id))
	query := `
		INSERT INTO users (auth_id, email, nickname, account_type_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (auth_id) DO UPDATE
		  SET email = EXCLUDED.email,
		      nickname = EXCLUDED.nickname,
		      account_type_id = EXCLUDED.account_type_id
		RETURNING user_id;
	`

	var userID int
	err := db.QueryRow(query, u.AuthID, u.Email, u.Name, u.AccountTypeID).Scan(&userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("DB insert/update error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

// getUserByAuthIDHandler handles GET /users/{auth_id}.
func getUserByAuthIDHandler(w http.ResponseWriter, r *http.Request) {

	setHandlerHeaders(w, r, "GET", "OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only GET is allowed on this endpoint.")
		return
	}

	// URL is /users/<auth_id>

	authID := r.URL.Path[len("/users/"):]
	if authID == "" {
		http.Error(w, "Missing auth_id in URL", http.StatusBadRequest)
		return
	}

	var u User
	query := `
		SELECT  auth_id, email, nickname, account_type_id
		FROM users
		WHERE auth_id = $1;
	`
	err := db.QueryRow(query, authID).Scan(
		&u.AuthID,
		&u.Email,
		&u.Name,
		&u.AccountTypeID,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("DB query error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

//select * from account_types where account_type_id = (select account_type_id from users where auth_id='google-oauth2|111634823248820523553')
