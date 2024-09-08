package main

import (
	"crystalsage/config"
	"crystalsage/internal"
	"crystalsage/internal/middlewares"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

func pop[T comparable](xs *[]T, index int) T {
	if index < 0 {
		index = len(*xs) + index
	}
	x := (*xs)[index]
	if index < len(*xs)-1 {
		*xs = append((*xs)[:index], (*xs)[index+1:]...)
	} else {
		*xs = (*xs)[:index]
	}
	return x
}

// ReadOrbConfig reads and parses a YAML file into the Config struct
func ReadOrbConfig(filename string) (*internal.OrbConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	var config internal.OrbConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return &config, nil
}

func StartOrb() {
	config, err := ReadOrbConfig(config.ENV.YAML_PATH)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
		return
	}
	internal.GlobalOrb = &internal.Orb{}
	internal.GlobalOrb.Load(*config)
}

func main() {
	fmt.Print("\033[H\033[2J")
	fmt.Println("[Starting Crystal Sage]")
	config.LoadEnvironment()
	fmt.Println("[Reading", config.ENV.YAML_PATH, "]")
	StartOrb()
	mux := http.NewServeMux()
	internal.GlobalOrb.Register(mux)
	port := internal.GlobalOrb.Port
	fmt.Println("Crystal Sage is running on port [", port, "]")
	loggingMiddleware := middlewares.Logging(mux)
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(port), loggingMiddleware))
}
