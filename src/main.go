// Package p contains an HTTP Cloud Function.
package main

import (
	"bytes"
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

func main() {
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

	var realData WeatherData
	json.Unmarshal([]byte(body), &realData)

	postData(averageTemp(realData), sum(realData, "rain"))
	getData()
}

func createClient(ctx context.Context) *firestore.Client {
	//init GCP firestore client
	sa := option.WithCredentialsFile("../gcloud-service-key.json")
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

	goodDays := 0

	for i := 0; i < 5; i++ {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed getting historical data: %v", err)
		}

		rainfall := doc.Data()["total-rainfall"]
		temp := doc.Data()["mean-temp"]
		if rainfall.(float64) >= 8.0 && temp.(float64) <= 12.0 {
			goodDays++
		}
	}

	if goodDays >= 3 {
		switch goodDays {
		case 3:
			message := []byte(`{"content":"3 days of good conditions in the last 5 days, dust off the deheydrator"}`)
			postDiscord(message)
		case 4:
			message := []byte(`{"content":"4 days of good conditions in the last 5 days, get on down there!"}`)
			postDiscord(message)
		case 5:
			message := []byte(`{"content":"5 days of good conditions in the last 5 days, WARNING WARNING EXTREMLY GOOD CONDITIONS, beware of coles bag bandits"}`)
			postDiscord(message)
		}
	}
}

func postDiscord(message []byte) {
	req, err := http.NewRequest("POST", "https://discord.com/api/webhooks/878448570531991573/Yru6y07_YdfwYtSf31R1Ya3jIib_X3t1q3Dfj3jco3X3yBLqZ7vaMMtUOF4R6s7UHLfb", bytes.NewBuffer(message))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
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
