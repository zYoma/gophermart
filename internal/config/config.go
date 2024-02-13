package config

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var flagAcrualtURL string
var flagLogLevel string
var flagDSN string
var flagTokenSecret string
var flagCheckOrderInterval int

const (
	envServerAddress = "RUN_ADDRESS"
	envAcrualURL     = "ACCRUAL_SYSTEM_ADDRESS"
	envLoggerLevel   = "LOG_LEVEL"
	envDSN           = "DATABASE_URI"
	envTokenSecret   = "TOKEN_SECRET"
	envOrderInterval = "CHECK_ORDER_INTERVAL"
)

type Config struct {
	RunAddr            string
	AcrualURL          string
	LogLevel           string
	DSN                string
	TokenSecret        string
	CheckOrderInterval int
}

func GetConfig() (*Config, error) {
	// парсим аргументы командной строки
	flag.StringVar(&flagRunAddr, "a", ":8081", "address and port to run server")
	flag.StringVar(&flagAcrualtURL, "r", "http://localhost:8080", "accrual system url")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.StringVar(&flagDSN, "d", "", "DB DSN")
	flag.StringVar(&flagTokenSecret, "s", "secret_for_test_only", "secret for jwt")
	flag.IntVar(&flagCheckOrderInterval, "i", 60, "interval in seconds between attempts to check the reason")
	flag.Parse()

	// если есть переменные окружения, используем их значения
	if envRunAddr := os.Getenv(envServerAddress); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envAcrualSystemURL := os.Getenv(envAcrualURL); envAcrualSystemURL != "" {
		flagAcrualtURL = envAcrualSystemURL
	}
	if envLogLevel := os.Getenv(envLoggerLevel); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
	if envDBDSN := os.Getenv(envDSN); envDBDSN != "" {
		flagDSN = envDBDSN
	}
	if envJWTSecret := os.Getenv(envTokenSecret); envJWTSecret != "" {
		flagTokenSecret = envJWTSecret
	}
	if envCheckOrderInterval := os.Getenv(envOrderInterval); envCheckOrderInterval != "" {
		intValue, err := strconv.Atoi(envCheckOrderInterval)
		if err != nil {
			return nil, err
		} else {
			flagCheckOrderInterval = intValue
		}
	}

	return &Config{
		RunAddr:            flagRunAddr,
		AcrualURL:          flagAcrualtURL,
		LogLevel:           flagLogLevel,
		DSN:                flagDSN,
		TokenSecret:        flagTokenSecret,
		CheckOrderInterval: flagCheckOrderInterval,
	}, nil
}
