package main

import (
	"context"
	"github.com/KushnerykPavel/integration/internal/app"

	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
)

func main() {
	var cfg app.Config
	err := envconfig.Process("consumer", &cfg)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	ctx := context.Background()

	os.Exit(app.New(&cfg).Run(ctx))
}
