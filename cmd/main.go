package main

import (
	"fmt"
	"log"

	"golang.org/x/sync/errgroup"
	router "wwwj.dev/wujiao/onlyoffice-simple-client/internal"
	"wwwj.dev/wujiao/onlyoffice-simple-client/internal/utils/config"
)

var (
	g errgroup.Group
)

func main() {
	r := router.InitRouter()

	g.Go(func() error {
		return r.Run(fmt.Sprintf(":%s", config.Conf.Server.ServerPort))
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
