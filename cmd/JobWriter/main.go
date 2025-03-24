package main

import (
	"github.com/hankimmy/PtmrBackend/pkg/google"
	"github.com/hankimmy/PtmrBackend/pkg/service"
	"github.com/rs/zerolog/log"

	"JobWriter/api"
)

func main() {
	dependencies, err := service.InitializeService()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}
	defer dependencies.StopFunc()
	gapi := google.NewGoogleService()
	server := api.NewServer(dependencies.Config, dependencies.ESClient, dependencies.TokenMaker, gapi)
	server.SetupRouter()
	err = server.Start(dependencies.Config.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
