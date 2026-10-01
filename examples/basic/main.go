package main

import (
	"fmt"
	"log"

	"example.com/my-app/internal/config"
	gocfg "github.com/tidjee-dev/gocfg"
)

func main() {
	if err := gocfg.LoadEnv(); err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.App.Name)
}
