package main

import (
	"context"
	"github.com/KushnerykPavel/api/internal/app"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
)

func main() {
	var cfg app.Config
	err := envconfig.Process("api", &cfg)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	ctx := context.Background()

	os.Exit(app.New(&cfg).Run(ctx))
}
