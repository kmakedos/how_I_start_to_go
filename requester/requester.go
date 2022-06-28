package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func request(c chan string, client *http.Client, url *string) {
	resp, err := client.Get(*url)
	if err != nil {
		// Connection Broken
		log.Fatalln(err.Error())
	}
	b, err := ioutil.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c <- err.Error()
		}
	}(resp.Body)
	if err != nil {
		c <- err.Error()
	}
	c <- string(b)
}

func main() {
	url := flag.String("u", "", "URL to send requests to")
	file := flag.String("f", "", "A file to read requests from")
	reqNumber := flag.Int("n", 1, "Number of requests to make")
	slowDown := flag.Duration("s", 100*time.Millisecond, "Duration to wait before each request, in time units eg. 1s or 100ms")
	flag.Parse()
	if *url == "" {
		log.Fatalln("Empty url given")
	}
	request_texts := []string{""}
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			request_texts = append(request_texts, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatalln(err)
		}
	}
	c := make(chan string)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	for i := 0; i < *reqNumber; i++ {
		if len(request_texts) > 1 {
			for _, text := range request_texts {
				requestURL := strings.Join([]string{*url, text}, "/")
				go request(c, client, &requestURL)
				time.Sleep(*slowDown)
				fmt.Printf("%s", <-c)
			}
		} else {
			go request(c, client, url)
			time.Sleep(*slowDown)
			fmt.Printf("%s", <-c)
		}
	}
}
