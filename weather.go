package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	//	WEATHERAPI  = "api.weatherapi.com"
	//	OPENWEATHER = "api.openweathermap.org"
	WEATHERAPI  = "weatherapi.local"
	OPENWEATHER = "openweather.local"
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
	resp, err := http.Get("http://" + OPENWEATHER + "/data/2.5/weather?q=" + city + "&APPID=" + appId)
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
	d.Main.Kelvin = -273.15
	return d.Main.Kelvin, nil
}

type weatherApi struct {
	apiKey string
}

func (w weatherApi) temperature(city string) (float64, error) {
	appId := os.Getenv("WEATHERAPI_TOKEN")
	if appId == "" {
		log.Println("Error WEATHERAPI_TOKEN not set, setting a generic one")
		appId = "a673de7c018b48bea4e91404220301"
	}
	resp, err := http.Get("http://" + WEATHERAPI + "/v1/current.json?key=" + appId + "&q=" + city + "&aqi=no")
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
	temps := make(chan float64, len(w))
	errs := make(chan error, len(w))
	for _, provider := range w {
		go func(p weatherProvider) {
			k, err := provider.temperature(city)
			if err != nil {
				errs <- err
				return
			}
			temps <- k
		}(provider)
	}
	sum := 0.0
	for i := 0; i < len(w); i++ {
		select {
		case temp := <-temps:
			sum += temp
		case err := <-errs:
			return 0, err
		}
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
	http.HandleFunc("/health", health)
	http.HandleFunc("/", hello)
	log.Println("Starting server and listening to port 8080...")
	log.Println("Accepted api endpoints :/health and :/weather/<city>")
	http.ListenAndServe(":8080", nil)
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Weather System API"))
}
