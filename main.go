package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type weatherProvider interface {
	temperature(city string) (float64, error)
}

type openWeatherMap struct{}

func (w openWeatherMap) temperature(city string) (float64, error) {
	appId := os.Getenv("OPENMAP_TOKEN")
	if appId == "" {
		log.Println("Error OPENMAP_TOKEN Not set")
	}
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + "&APPID=" + appId)
	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
		Name string `json:"name"`
	}
	if resp.StatusCode != 200 {
		log.Println("Error " + resp.Status)
		return 0, err
	}
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}
	d.Main.Kelvin -= 273.15
	return d.Main.Kelvin, nil
}

type weatherApi struct {
	apiKey string
}

func (w weatherApi) temperature(city string) (float64, error) {
	appId := os.Getenv("WEATHERAPI_TOKEN")
	if appId == "" {
		log.Println("Error WEATHERAPI_TOKEN Not set")
	}
	resp, err := http.Get("http://api.weatherapi.com/v1/current.json?key=" + appId + "&q=" + city + "&aqi=no")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var d struct {
		Observation struct {
			Celsius float64 `json:"temp_c"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}
	return d.Observation.Celsius, nil
}

type multiWeatherProvider []weatherProvider

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	sum := 0.0
	for _, provider := range w {
		k, err := provider.temperature(city)
		if err != nil {
			return 0, err
		}
		sum += k
	}
	return sum / float64(len(w)), nil
}

func main() {

	mw := multiWeatherProvider{
		openWeatherMap{},
		weatherApi{apiKey: ""},
	}

	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		city := strings.SplitN(r.URL.Path, "/", 3)[2]
		log.Printf("Querying city %s\n", city)
		temp, err := mw.temperature(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"city": city,
			"temp": temp,
			"took": time.Since(begin).String(),
		})
	})
	http.ListenAndServe(":8080", nil)
	http.HandleFunc("/", hello)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello!"))
}
