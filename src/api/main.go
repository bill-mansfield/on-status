// Package p contains an HTTP Cloud Function.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func createClient(ctx context.Context) *firestore.Client {
	sa := option.WithCredentialsFile("gcloud-service-key.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	return client
}

func postData() {
	// Get a Firestore client.
	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	_, _, err := client.Collection("Historical data").Add(ctx, map[string]interface{}{
		"mean-temp":      "11",
		"total-rainfall": "0.00",
	})
	if err != nil {
		log.Fatalf("Failed adding historical data: %v", err)
	}

}

func main() {
	// Get data from BOM
	resp, err := http.Get("http://www.bom.gov.au/fwo/IDS60801/IDS60801.94811.json")
	if err != nil {
		log.Fatalln(err)
	}

	// Response is a 403, the BOM are actively disallowing "webscraping" of their .json endpoint for weather data..."
	// Attempt to replicate browser request to skirt this issue

	fmt.Println(resp)

	fmt.Println("Task complete")
}
