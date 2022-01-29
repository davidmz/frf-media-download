package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/golang-jwt/jwt"
	"gopkg.in/ini.v1"
)

type Config struct {
	Token      string
	ApiRoot    string
	ResultsDir string
	Threads    int

	resultsRoot string
	apiRootURL  *url.URL
}

var config = &Config{
	ApiRoot:    "https://freefeed.net",
	ResultsDir: "./results",
	Threads:    10,
}

func mustLoadConfig(fileName string) {
	iniData, err := ini.Load(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load config file: %v\n", err)
		os.Exit(1)
	}

	if err := iniData.MapTo(config); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %v\n", err)
		os.Exit(1)
	}

	_, _, err = new(jwt.Parser).ParseUnverified(config.Token, new(jwt.StandardClaims))
	if err != nil {
		fmt.Fprintf(os.Stderr, "The access token is invalid. Please generate the access token using instructions from the config file.\n")
		os.Exit(1)
	}

	config.apiRootURL, err = url.Parse(config.ApiRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid server URL: ", err)
		os.Exit(1)
	}
}
