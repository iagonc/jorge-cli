package main

import (
	"github.com/iagonc/jorge-cli/config"
	"github.com/iagonc/jorge-cli/router"
)

func main() {
	// Init config
	logger := config.GetLogger()
	
	err:=config.Init()
	
	if err != nil {
		logger.Error("test")
		return
	}

	// Init Router
	router.Initialize()
}
