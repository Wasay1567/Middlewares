package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Post struct {
	Title      string
	Created_at string
	Likes      int
}

const PrivateKey = "htm23Cv56"

var posts = []Post{}

func getAllPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(posts)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var payload Post
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid Post Format", http.StatusBadRequest)
		return
	}
	p := Post{
		Title:      payload.Title,
		Created_at: time.Now().Format("Jan 2 15:04"),
		Likes:      payload.Likes,
	}
	posts = append(posts, p)
	json.NewEncoder(w).Encode(p)
	w.WriteHeader(http.StatusOK)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(st)
		log.Printf("Completed request: %s %s in %v\n", r.Method, r.URL.Path, duration)
	})
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != PrivateKey {
			http.Error(w, "Unauthorized User", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload Credentials
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.Username != "Wasay123" && payload.Password != "Wasay786@" {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	token, err := generateJWT(payload.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
	w.WriteHeader(http.StatusOK)
}

func jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token_string := r.Header.Get("Authorization")

		if token_string == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(token_string, &claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		username := claims["username"].(string)
		log.Printf("Authenticated User: %v", username)
		next.ServeHTTP(w, r)
	})
}

func main() {

	mux := http.NewServeMux()

	// Routes
	mux.Handle("GET /posts", loggingMiddleware(http.HandlerFunc(getAllPost)))
	mux.Handle("POST /posts", loggingMiddleware(jwtAuthMiddleware(http.HandlerFunc(createPost))))
	mux.Handle("POST /account", loggingMiddleware(http.HandlerFunc(LoginHandler)))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	srv.ListenAndServe()
	fmt.Println("Server Started Successfully...")
}
