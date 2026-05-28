package main

import (
	"fmt"

	llmclient "github.com/victoria-hft/go-llm-client"
)

func main() {
	fmt.Printf("go-llm-client version: %s\n", llmclient.Version())
}
