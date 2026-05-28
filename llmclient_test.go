package llmclient_test

import (
	"testing"

	llmclient "github.com/victoria-hft/go-llm-client"
)

func TestVersionIsAvailableToImporters(t *testing.T) {
	if llmclient.Version() == "" {
		t.Fatal("Version() returned an empty string")
	}
}
