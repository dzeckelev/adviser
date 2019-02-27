package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type inputResp []*inputItem
type outputResp []*outputItem

type inputItem struct {
	IndexStrings    []string           `json:"index_strings"`
	CountryCode     string             `json:"country_code"`
	StateCode       interface{}        `json:"state_code"` // unknown type
	Cases           map[string]string  `json:"cases"`
	Coordinates     map[string]float64 `json:"coordinates"`
	CountryCases    interface{}        `json:"country_cases"` // unknown type
	Code            string             `json:"code"`
	Name            string             `json:"name"`
	Weight          int64              `json:"weight"`
	Type            string             `json:"type"`
	CountryName     string             `json:"country_name"`
	MainAirportName interface{}        `json:"main_airport_name"` // unknown type
}

// DstItem is a output item.
type outputItem struct {
	Slug     string `json:"slug"`
	Subtitle string `json:"subtitle"`
	Title    string `json:"title"`
}

type server struct {
	reqTimeout    time.Duration
	httpServer    *http.Server
	targetAddress string
}

// config is an application configuration.
type config struct {
	RequestTimeout uint64 // In milliseconds.
	Addr           string
	TargetAddr     string
}

// newConfig creates a new default config.
func newConfig() *config {
	return &config{
		Addr:           ":80",
		RequestTimeout: 3000,
		TargetAddr:     "https://places.aviasales.ru",
	}
}

func readConfig(name string, data interface{}) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(data)
}

func request(url string, timeout time.Duration, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("internal error")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &result)
}

func newServer(cfg *config) *server {
	srv := &server{
		reqTimeout:    time.Duration(cfg.RequestTimeout) * time.Millisecond,
		httpServer:    &http.Server{Addr: cfg.Addr},
		targetAddress: cfg.TargetAddr,
	}

	srv.httpServer.Handler = http.HandlerFunc(srv.handlerFunc)

	return srv
}

func (s *server) listenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *server) handlerFunc(w http.ResponseWriter, r *http.Request) {
	url := s.targetAddress + r.URL.String()

	input := inputResp{}
	if err := request(url, s.reqTimeout*time.Millisecond, &input); err != nil {
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	output := make(outputResp, len(input))
	for k := range input {
		// code -> slug
		// country_name -> subtitle
		// name -> title
		item := &outputItem{
			Slug:     input[k].Code,
			Subtitle: input[k].CountryName,
			Title:    input[k].Name,
		}

		output[k] = item
	}

	result, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func main() {
	cfg := newConfig()
	fConfig := flag.String("config", "config.json", "Configuration file")

	flag.Parse()

	if err := readConfig(*fConfig, cfg); err != nil {
		panic(err)
	}

	err := newServer(cfg).listenAndServe()
	if err != nil {
		panic(err)
	}
}
