package main

import (
	// "basic-web-server/api"
	// "fmt"
	"net/http"

	"encoding/json"
	"fmt"
	"log"
	"net"
	// "net/http"
	"os"
	// "strings"
	"github.com/joho/godotenv"
	"github.com/oschwald/geoip2-golang"
	// "github.com/gorilla/mux"
)

// func Handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "<h1>your successfully deployed golang page</h1>")
// }

// func Handle(w http.ResponseWriter, r *http.Request) {
// 	// server := api.NewServer()
// 	// server.Router.ServeHTTP(w, r)
// }

type Response struct {
	ClientIP string `json:"client_ip"`
	Location string `json:"location"`
	Greeting string `json:"greeting"`
}

// func getIP(r *http.Request) string {
// 	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
// 	userIP := net.ParseIP(ip)
// 	if userIP == nil {
// 		userIP = net.ParseIP(r.Header.Get("X-Real-Ip"))
// 	}
// 	if userIP == nil {
// 		userIP = net.ParseIP(r.Header.Get("X-Forwarded-For"))
// 	}
// 	if userIP == nil {
// 		userIP = net.IPv4(127, 0, 0, 1) // Default to localhost
// 	}
// 	return userIP.String()
// }

func getIP(r *http.Request) string {
	// Try to get the IP address from the X-Real-Ip header
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		// If X-Real-Ip is empty, try the X-Forwarded-For header
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		// If both headers are empty, fall back to RemoteAddr
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	// Parse the IP address
	userIP := net.ParseIP(ip)

	// If parsing failed or no IP was found, default to localhost
	if userIP == nil {
		userIP = net.IPv4(127, 0, 0, 1)
	}

	return userIP.String()
}

func getLocation(ip string) (string, string) {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	parsedIP := net.ParseIP(ip)
	record, err := db.City(parsedIP)
	if err != nil {
		log.Fatal(err)
	}

	return record.City.Names["en"], record.Country.Names["en"]
}

func getTemperature(city string) (float64, error) {

	if city == "" {
		return 0, fmt.Errorf("city not provided")
	}


	if _, exists := os.LookupEnv("RAILWAY_ENVIRONMENT"); exists == false {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("API key not set")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", city, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received non-200 response: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var weatherResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		return 0, fmt.Errorf("error decoding JSON response: %v", err)
	}

	// Extract the temperature value
	main, ok := weatherResponse["main"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("error parsing 'main' field from response")
	}

	temp, ok := main["temp"].(float64)
	if !ok {
		return 0, fmt.Errorf("error parsing 'temp' field from response")
	}

	return temp, nil
}
// test
func helloHandler(w http.ResponseWriter, r *http.Request) {
	visitorName := r.URL.Query().Get("visitor_name")
	clientIP := getIP(r)
	city, _ := getLocation(clientIP)

	// temp, err := getTemperature(url.QueryEscape(city))
	// if err != nil {
	// 	log.Printf("Error getting temperature: %v", err)
	// 	http.Error(w, "Could not fetch temperature", http.StatusInternalServerError)
	// 	return
	// }

	greeting := fmt.Sprintf("Hello, %s!, the temperature is %s degrees Celsius in %s", visitorName, city)

	response := Response{
		ClientIP: clientIP,
		Location: city,
		Greeting: greeting,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func main() {

	// Hello world http server
	// http.HandleFunc("/hello-world", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("hello world"))
	// })
	// http.ListenAndServe(":8080", nil)

	http.HandleFunc("/api/hello", helloHandler)
	// srv:= api.NewServer()
	// r := mux.NewRouter()
	// r.HandleFunc("/", Handler).Methods("GET")
	// http.ListenAndServe(":8080", srv)
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
