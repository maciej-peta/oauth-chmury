package main

import (
	"database/sql"
	"fmt"
	"github.com/MicahParks/keyfunc"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	kilobyte = 1024
	megabyte = kilobyte * 1024
)

func main() {
	ServerAddress := ":" + os.Getenv("PORT")

	//loop to connect to postgres - needed because otherwise golang can race ahead of the db
	var driverErr error
	var pingErr error
	dbInfo := os.Getenv("DATABASE_INFO")
	attemptLimit := 10

	for i := 1; i < attemptLimit; i++ {
		db, driverErr = sql.Open("postgres", dbInfo)
		if driverErr == nil {
			if pingErr = db.Ping(); pingErr == nil {
				driverErr = nil
				pingErr = nil
				break
			}
		}
		fmt.Println("Server to database connection failed. Retrying...")
		time.Sleep(time.Millisecond * 200)
	}

	if driverErr != nil {
		log.Fatal(driverErr)
	}
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Server connected to Postgres.")

	//mux

	//wanted to make handlers goroutines, but accorting to this:
	//https://eli.thegreenplace.net/2021/life-of-an-http-request-in-a-go-server/
	//the net/http mux already creates new goroutines when called.

	mux := http.NewServeMux()

	mux.HandleFunc("/jpeg/png", imageHandlerFactory(jpegTag, pngTag))
	mux.HandleFunc("/jpeg/webp", imageHandlerFactory(jpegTag, webpTag))

	mux.HandleFunc("/png/jpeg", imageHandlerFactory(pngTag, jpegTag))
	mux.HandleFunc("/png/webp", imageHandlerFactory(pngTag, webpTag))

	mux.HandleFunc("/webp/jpeg", imageHandlerFactory(webpTag, jpegTag))
	mux.HandleFunc("/webp/png", imageHandlerFactory(webpTag, pngTag))

	mux.HandleFunc("/users", createOrUpdateUserHandler)
	mux.HandleFunc("/users/", getUserByAuthIDHandler)

	mux.HandleFunc("/health", healthHandler)

	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", auth0Domain)
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		log.Fatalf("Failed to get JWKS: %v", err)
	}
	handler := jwtMiddleware(mux, jwks)

	log.Println("Backend server is running on", ServerAddress)
	if err := http.ListenAndServe(ServerAddress, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	} else {
		log.Printf("Server started")
	}
}

//todo: verify limits of requests

//todo: add a timeout for data transfer into the server

//seems its ok and if i do these 2 ^ it should be ok to send in
