// Package p contains an HTTP Cloud Function.
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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
	// Get data from Open Weather

	// The api url below contains the following prarams
	// Latitude
	// Longitude
	// dt: unix time stamp from which to get weather data from the previous 24hrs
	// Api key
	// If we call this end point once a day we can collate the previous 24hrs of weather everyday
	// Creating our own free historical weather data

	// The data is returned in hourly blocks,
	// We must add up rainfall if present
	// We must average temp

	// Step 1: Get data from Open Weather api on a daily basis via github action
	// Step 2: Add up and average data -> save to DB
	// Step 3: Read data from DB compare it to model for perfect conditions
	// Step 4: If conditions are perfect notify discord

	currentTime := time.Now().Unix()
	fmt.Println("Current Unix Time:", currentTime)

	resp, err := http.Get("https://api.openweathermap.org/data/2.5/onecall/timemachine?lat=-35.55&lon=138.2333&dt=1629681702&appid=9219e14c1a190adc052618d596aa7e28")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bs := string(body)

	fmt.Println(bs)

	fmt.Println(resp)

	fmt.Println("Task complete")
}
