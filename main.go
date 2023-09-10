package main

import (
	"fmt"
	"log"
	"net/http"

	"git.ekzyis.com/ekzyis/delphi.market/env"
	"git.ekzyis.com/ekzyis/delphi.market/server"
)

var (
	s *server.Server
)

func init() {
	log.Printf("Commit:      %s", env.CommitShortSha)
	log.Printf("Public URL:  %s", env.PublicURL)
	log.Printf("Environment: %s", env.Env)

	s = server.NewServer()
}

func main() {
	if err := s.Start(fmt.Sprintf("%s:%d", "127.0.0.1", env.Port)); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
