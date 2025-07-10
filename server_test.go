package main

import (
	"testing"

	"github.com/suaobra/suaobra-app/server"
)

func TestDiscord(t *testing.T) {
	server.NotifyDiscord("test")
}
