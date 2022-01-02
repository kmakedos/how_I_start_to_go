package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type weatherData struct {
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
	Name string `json:"name"`
}

func query(city string) (weatherData, error) {
	appId := os.Getenv("OPENMAP_TOKEN")
	if appId == "" {
		log.Println("Error OPENMAP_TOKEN Not set")
	}
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + "&APPID=" + appId)
	var d weatherData
	if resp.StatusCode != 200 {
		log.Println("Error " + resp.Status)
		return d, err
	}
	if err != nil {
		return weatherData{}, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}
	d.Main.Kelvin -= 273.15
	return d, nil
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		city := strings.SplitN(r.URL.Path, "/", 3)[2]
		log.Printf("Querying city %s\n", city)
		data, err := query(city)
		log.Println("Data analyzed:")
		log.Println(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		json.NewEncoder(w).Encode(data)

	})
	http.ListenAndServe(":8080", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello!"))
}
