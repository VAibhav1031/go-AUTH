package main

import (
	"github.com/VAibhav1031/go-AUTH/internal/cli"
	"github.com/VAibhav1031/go-AUTH/internal/logger"
)

func main() {
	logger.LoggerInitiator()
	cli.Initiate()

}
