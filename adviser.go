package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	applicationJSON = "application/json"
	contentType     = "Content-Type"
	processingTime  = "processing time"
	request         = "request"
	response        = "response"
	urlStr          = "url"
)

// internalErr is a standard error.
var internalErr = []byte(`{"error": "internal error"}`)

// inputResp is a target service response structure.
type inputResp []*inputItem

// inputResp is a response that a client expects.
type outputResp []*outputItem

// outputItem is a input item.
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

// outputItem is a output item.
type outputItem struct {
	Slug     string `json:"slug"`
	Subtitle string `json:"subtitle"`
	Title    string `json:"title"`
}

// server is a http server.
type server struct {
	debug         bool
	httpServer    *http.Server
	logger        *zap.SugaredLogger
	reqTimeout    time.Duration
	targetAddress string
}

// config is an application configuration.
type config struct {
	Addr           string
	LogLevel       string
	RequestTimeout uint64 // In milliseconds.
	TargetAddr     string
}

// newConfig creates a new default config.
func newConfig() *config {
	return &config{
		Addr:           ":80",
		LogLevel:       "info",
		RequestTimeout: 3000,
		TargetAddr:     "https://places.aviasales.ru",
	}
}

// readConfig reads configuration from file.
func readConfig(name string, data interface{}) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	return json.NewDecoder(file).Decode(data)
}

func (s *server) request(logger *zap.SugaredLogger, url string,
	timeout time.Duration, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Error(err)
		return err
	}

	req = req.WithContext(ctx)
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf(http.StatusText(res.StatusCode))
		logger.Error(err)
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func newServer(logger *zap.SugaredLogger, cfg *config) *server {
	var debug bool

	if cfg.LogLevel == "debug" || cfg.LogLevel == "DEBUG" {
		debug = true
	}

	srv := &server{
		debug:         debug,
		httpServer:    &http.Server{Addr: cfg.Addr},
		logger:        logger,
		reqTimeout:    time.Duration(cfg.RequestTimeout) * time.Millisecond,
		targetAddress: cfg.TargetAddr,
	}

	srv.httpServer.Handler = http.HandlerFunc(srv.handlerFunc)

	return srv
}

func (s *server) listenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *server) addJSONContentType(w http.ResponseWriter) {
	w.Header().Add(contentType, applicationJSON)
}

// handlerFunc processes requests.
func (s *server) handlerFunc(w http.ResponseWriter, r *http.Request) {
	// Makes target url.
	url := s.targetAddress + r.URL.String()
	logger := s.logger.With(urlStr, url)

	// Adds "Content-Type" = "application/json".
	s.addJSONContentType(w)

	// Logs the time to process a request.
	if s.debug {
		start := time.Now()
		defer func() {
			stop := time.Now()

			logger.With(processingTime,
				stop.Sub(start).String()).Debug(request)
		}()
	}

	// Receives a response from the target service.
	input := inputResp{}
	if err := s.request(logger, url,
		s.reqTimeout*time.Millisecond, &input); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(internalErr); err != nil {
			logger.Error(err)
		}
		return
	}

	logger = logger.With(response, input)

	// Makes a result.
	output := make(outputResp, len(input))
	for k := range input {
		// code -> slug
		// country_name -> subtitle
		// name -> title
		output[k] = &outputItem{
			Slug:     input[k].Code,
			Subtitle: input[k].CountryName,
			Title:    input[k].Name,
		}
	}

	result, err := json.Marshal(output)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(result); err != nil {
		logger.Error(err)
	}
}

func newLogger(level string) *zap.SugaredLogger {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))

	logger, _ := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()

	return logger.Sugar()
}

func main() {
	fConfig := flag.String("config", "config.json", "Configuration file")

	flag.Parse()

	cfg := newConfig()
	if err := readConfig(*fConfig, cfg); err != nil {
		panic(err)
	}

	logger := newLogger(cfg.LogLevel)
	defer func() {
		_ = logger.Sync()
	}()

	logger = logger.With("config", cfg)

	err := newServer(logger, cfg).listenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
