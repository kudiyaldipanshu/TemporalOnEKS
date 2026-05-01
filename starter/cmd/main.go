package main

import (
	"log"
	"net/http"
	"os"
	"temporal/starter/api"

	"go.temporal.io/sdk/client"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func main() {
	host := os.Getenv("TEMPORAL_HOST")
	if host == "" {
		host = "temporal:7233"
	}
	c, err := client.Dial(client.Options{
		HostPort: host,
	})

	if err != nil {
		log.Fatalf("failed to connect to Temporal: %v", err)
	}

	defer c.Close()

	handler := api.NewHandler(c)

	http.HandleFunc("/order", handler.CreateOrder)
	http.HandleFunc("/health", healthHandler)

	log.Println("Starter API running on port 8080")

	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatalln("Server failed: ", err)
	}
}
