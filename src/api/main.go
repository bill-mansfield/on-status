// Package p contains an HTTP Cloud Function.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Rain struct {
	OneHour float64 `json:"1h"`
}

type Hourly struct {
	Temp float64
	Rain Rain
}

type WeatherData struct {
	Hourly []Hourly
}

func createClient(ctx context.Context) *firestore.Client {
	sa := option.WithCredentialsFile("../../gcloud-service-key.json")
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

func postData(temp float64, rain float64) {
	// Get a Firestore client.
	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	date := time.Now().Format("2006/01/02")
	unixDate := time.Now().Unix()
	// Post data
	_, _, err := client.Collection("Historical data").Add(ctx, map[string]interface{}{
		"mean-temp":      temp,
		"total-rainfall": rain,
		"date":           date,
		"unixDate":       unixDate,
	})
	if err != nil {
		log.Fatalf("Failed adding historical data: %v", err)
	}
}

func getData() {
	// Get a Firestore client.
	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	// Get previous 5 days of data
	iter := client.Collection("Historical data").Where("unixDate", ">=", time.Now().Unix()-432000).Documents(ctx)
	defer iter.Stop()

	for i := 0; i < 5; i++ {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed getting historical data: %v", err)
		}

		// 5 docs are printed here
		// While we are interating through the data here, determine if the conditions are good then call the discord webhook
		fmt.Println(doc.Data())
	}
}

func main() {
	// postData()
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

	// Now minus 24hrs
	twentyFourHoursAgo := time.Now().Unix() - 86400
	apiKey := "9219e14c1a190adc052618d596aa7e28"
	endpoint := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall/timemachine?lat=-35.55&lon=138.2333&dt=%d&appid=%s", twentyFourHoursAgo, apiKey)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Printf("%T\n", body)

	var realData WeatherData
	json.Unmarshal([]byte(body), &realData)

	fmt.Print("Total rainfall: ", sum(realData, "rain"))
	fmt.Printf("Average temp: %.2f", averageTemp(realData))

	// postData(averageTemp(realData), sum(realData, "rain"))

	getStatus()
}

func getStatus() {

	getData()
}

func sum(data WeatherData, dataType string) float64 {
	// Sum hourly rainfall or temp data to create daily record

	result := 0.0

	for _, hour := range data.Hourly {
		if dataType == "rain" {
			result += hour.Rain.OneHour
		} else if dataType == "temp" {
			result += hour.Temp
		} else {
			log.Fatalln("Provide rain or temp for dataType arg")
		}
	}

	result = math.Floor(result*100) / 100
	return result
}

func averageTemp(data WeatherData) float64 {
	// Get average temp convert from kelvin to celsius

	result := (sum(data, "temp") / 24) - 273.15
	return math.Floor(result*100) / 100
}
