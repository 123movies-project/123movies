package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

/*
func main(){

	if canIHaveAProxy() {
		log.Println("Candy shop open")
	}
}
*/

func canIHaveAProxy() bool {
	ipapiClient := http.Client{}
	req, err := http.NewRequest("GET", "https://ipapi.co/json/", nil)
	if err != nil {
		log.Println(err)
		log.Println("Error 116")
		return false
	}
	req.Header.Set("User-Agent", "ipapi.co/#go-v1.3")
	resp, err := ipapiClient.Do(req)
	if err != nil {
		log.Println(err)
		log.Println("Error 116")
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Println(err)
		log.Println("Error 116")
		return false
	}
	fmt.Println(string(body))
	respString := string(body)
	NoCandyArr := []string{"Euro", "United States", "United Kingdom", "Canada", "Australia", "New Zealand", "Ireland", "Isreal"}
	for _, nocandyString := range NoCandyArr {
		if strings.Contains(respString, nocandyString) {
			log.Println("Error 117")
			return false
		}
	}
	return true
}
