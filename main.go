package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type weatherData struct {
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
	Name string `json:"name"`
}

func query(city string) (weatherData, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city + "&APPID=f43f1ddb6354407ef500954130aca9bd")
	if resp.StatusCode != 200 {
		log.Println("Error " + resp.Status)
	}
	if err != nil {
		return weatherData{}, err
	}
	defer resp.Body.Close()

	var d weatherData
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
