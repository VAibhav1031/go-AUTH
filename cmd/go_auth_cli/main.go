package main

import (
	"fmt"

	"github.com/VAibhav1031/go-AUTH/internal/cli"
	"github.com/VAibhav1031/go-AUTH/internal/logger"
)

func main() {
	fmt.Println("***********GO-CLI-AUTH***********")
	logger.LoggerInitiator()
	cli.Initiate()

}
