package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Name      string      `json:"name"`
	Character []Character `json:"character"`
}

type Character struct {
	Name  string `json:"name"`
	Power int    `json:"max_power"`
}

var (
	store = make(map[string]int)
	usage = make(map[string]int)

	avengers = "http://www.mocky.io/v2/5ecfd5dc3200006200e3d64b"
	mutants  = "http://www.mocky.io/v2/5ecfd6473200009dc1e3d64e"
	anti     = "https://run.mocky.io/v3/e5ab003a-4919-4e2d-ad06-247ff1a6e558"
	arr      = [3]string{avengers, mutants, anti}
)

func request(url string) []Character {
	spaceClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	responseObject := Response{}
	jsonErr := json.Unmarshal(body, &responseObject)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	tempChars := responseObject.Character
	return tempChars
}

func addToStore(name string, power int) {
	_, usageprs := usage[name]
	if !usageprs {
		usage[name] = 0
	}

	_, storeprs := store[name]
	if !storeprs {
		for len(store) > 14 {
			deleteleastused()
		}
	}
	store[name] = power

}

func storeRequest(url string) {
	Chars := request(url)
	for i := 0; i < len(Chars); i++ {
		addToStore(Chars[i].Name, Chars[i].Power)
	}
}

func deleteweakest() {
	minpower := 100000000
	todelete := ""
	for name, power := range store {
		if power < minpower {
			todelete = name
			minpower = power
		}
	}
	delete(store, todelete)
	delete(usage, todelete)
}

func deleteleastused() { //which char is called the least
	minusage := 10000000
	todelete := ""
	for name, usage := range usage {
		if usage < minusage {
			todelete = name
			minusage = usage
		}
	}
	delete(store, todelete)
	delete(usage, todelete)
}

func tryfetch(tosearch string) string {
	pow := "Not Found"
	name := ""
	for _, url := range arr {
		chars := request(url)
		for _, char := range chars {
			if char.Name == tosearch {
				pow = fmt.Sprint(char.Power)
				name = char.Name
			}
		}
	}

	_, prs := store[name]
	if !prs {
		if pow != "Not Found" {
			for len(store) > 14 {
				deleteweakest()
			}
			i, _ := strconv.Atoi(pow)
			addToStore(name, i)
		}
	}

	return pow
}

func handler(w http.ResponseWriter, req *http.Request) {
	hero := req.URL.Query().Get("hero")
	fmt.Println(len(store))
	_, prs := store[hero]
	if prs {
		fmt.Fprintf(w, "%v\n", store[hero])
		usage[hero] += 1
		fmt.Println(usage)
	} else {
		fmt.Fprintf(w, "%v\n", tryfetch(hero))
	}
}

func repeater() { //power increment every 10s
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		for _, url := range arr {
			go storeRequest(url)
		}
	}
}

func main() {
	for _, url := range arr {
		storeRequest(url)
	}
	fmt.Println(store)
	go repeater()

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8081", nil)

	//http://localhost:8081/?hero=Apocalypse
}
