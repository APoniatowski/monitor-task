package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.54 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.54 Safari/537.36",
	// ... (add more user agents if you'd like to check broswer compatibility)
}

const (
	additionaTask = "http://api.nbp.pl/api/exchangerates/rates/a/pln/last/100/?format=json" // does not work, getting 404 on this
	hostURL       = "http://api.nbp.pl/api/exchangerates/rates/a/eur/last/100/?format=json"
	X             = 10 // Number of checks
	Y             = 5  // Time interval between checks in seconds
	logFile       = "log.txt"
)

func main() {
	for i := 0; i < X; i++ {
		fmt.Printf("Check %d:\n", i+1)
		logMessage := processCheck()
		fmt.Println(logMessage)

		// Append log message to log.txt
		appendLog(logMessage)

		if i < X-1 {
			time.Sleep(time.Second * Y)
		}
	}
}

func processCheck() string {
	startTime := time.Now()

	// Create a custom HTTP client
	client := &http.Client{}

	req, err := http.NewRequest("GET", hostURL, nil)
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	// Set random User-Agent header
	randomUserAgent := getRandomUserAgent()

	// Set User-Agent header
	req.Header.Set("User-Agent", randomUserAgent)

	// Send GET request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error sending GET request: %v", err)
	}
	defer resp.Body.Close()

	// Measure time taken
	elapsedTime := time.Since(startTime)

	// Check response code
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Response status code not OK: %d", resp.StatusCode)
	}

	// Check Content-Type (with charset)
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return fmt.Sprintf("Invalid Content-Type: %s", contentType)
	}

	// Decode JSON response
	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Sprintf("Error decoding JSON: %v", err)
	}

	// Validate JSON syntax
	if _, ok := data["rates"]; !ok {
		return "JSON syntax validation failed"
	}

	// Additional task: Check EUR prices against PLN within given ranges
	eurRates := data["rates"].([]interface{})
	for _, rate := range eurRates {
		rateMap := rate.(map[string]interface{})
		if currency, ok := rateMap["currency"]; ok && currency == "EUR" {
			if mid, ok := rateMap["mid"].(float64); ok {
				if mid < 4.5 || mid > 4.7 {
					date := rateMap["effectiveDate"].(string)
					fmt.Printf("EUR rate on %s is not in the range (PLN 4.5 - PLN 4.7)\n", date)
				}
			}
		}
	}

	// Build log message
	logMessage := fmt.Sprintf("Time: %s, Response Time: %s, Status Code: %d, Content-Type: %s",
		time.Now().Format(time.RFC3339), elapsedTime, resp.StatusCode, contentType)

	return logMessage
}

func appendLog(message string) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(message + "\n")
	if err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

func getRandomUserAgent() string {
	randIndex := rand.Intn(len(userAgents))
	return userAgents[randIndex]
}
