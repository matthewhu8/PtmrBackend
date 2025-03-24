package main

import (
	"github.com/hankimmy/PtmrBackend/pkg/service"
	"github.com/rs/zerolog/log"

	"MatchingService/api"
)

func main() {
	dependencies, err := service.InitializeService()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}
	server := api.NewServer(dependencies.Config, dependencies.ESClient, dependencies.TokenMaker, nil)
	server.SetupRouter()
	err = server.Start(dependencies.Config.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
