package main

import (
	"context"
	"fmt"
	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

var (
	auth0Domain   = "file-conversion-tenant.eu.auth0.com"
	auth0Audience = "https://file-conversion-api/"
)

const userAuthIDKey string = "userAuthID"

func jwtMiddleware(next http.Handler, jwks *keyfunc.JWKS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//i skip the auth middleware if somebody uses /health so that i dont
		//have to generate tokens for healthchecks.
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		setHandlerHeaders(w, r, "GET", "POST", "OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		audClaim, ok := claims["aud"]
		if !ok {
			http.Error(w, "Missing audience claim", http.StatusUnauthorized)
			return
		}

		audValid := false
		switch aud := audClaim.(type) {
		case string:
			audValid = aud == auth0Audience
		case []interface{}:
			for _, a := range aud {
				if aStr, ok := a.(string); ok && aStr == auth0Audience {
					audValid = true
					break
				}
			}
		}

		if !audValid {
			http.Error(w, "Invalid audience", http.StatusUnauthorized)
			return
		}

		expectedIss := fmt.Sprintf("https://%s/", auth0Domain)
		if claims["iss"] != expectedIss {
			http.Error(w, "Invalid issuer", http.StatusUnauthorized)
			return
		}
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			http.Error(w, "Missing sub claim", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userAuthIDKey, sub)

		//fmt.Println("user sub:", sub)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserAuthID(r *http.Request) (string, bool) {
	sub, ok := r.Context().Value(userAuthIDKey).(string)
	return sub, ok
}
