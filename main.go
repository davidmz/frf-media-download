package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("--------------------")
	fmt.Println("⏬ FrF Media Download")
	fmt.Println("🔗 github.com/davidmz/frf-media-download")
	fmt.Println("--------------------")
	fmt.Println("")

	var (
		showHelp   = flag.Bool("help", false, "Show this help message")
		configFile = flag.String("config", "./config.ini", "Path to the 'config.ini' file")
	)
	flag.Parse()

	if *showHelp {
		fmt.Fprintf(flag.CommandLine.Output(), "Flags of %s:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(0)
	}

	mustLoadConfig(*configFile)

	username, err := checkToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("👋 Hello, " + username + "!")
	fmt.Println("Starting download process...")
	fmt.Println("")

	config.resultsRoot = filepath.Join(config.ResultsDir, config.apiRootURL.Hostname(), username)

	if err := os.MkdirAll(config.resultsRoot, 0777); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create results directory: ", err)
		os.Exit(1)
	}

	if err := startDownload(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
