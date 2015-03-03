package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/rancherio/host-api/app/common"
	"github.com/rancherio/host-api/auth"
	"github.com/rancherio/host-api/config"
	"github.com/rancherio/host-proxy/proxy"
)

func main() {
	app := cli.NewApp()
	app.Email = ""
	app.Author = "Rancher Labs, Inc."
	app.Action = runApp
	app.Version = ""

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "public-key",
			Value:  "/var/lib/cattle/etc/cattle/api.crt",
			Usage:  "Public Key for Authentication",
			EnvVar: "CATTLE_HOST_PROXY_PUBLIC_KEY",
		},
		cli.StringFlag{
			Name:   "listen-ip",
			Value:  "0.0.0.0",
			Usage:  "Listen IP",
			EnvVar: "CATTLE_HOST_PROXY_LISTEN_IP",
		},
		cli.IntFlag{
			Name:   "listen-port",
			Value:  9345,
			Usage:  "Listen port",
			EnvVar: "CATTLE_HOST_PROXY_LISTEN_PORT",
		},
	}

	app.Run(os.Args)
}

func runApp(c *cli.Context) {
	config.Config.Auth = true
	config.Config.Key = c.String("public-key")

	if err := config.ParsedPublicKey(); err != nil {
		log.Fatalf("Failed to read public-key %s : %v", c.String("public-key"), err)
	}

	router := mux.NewRouter()

	http.Handle("/", auth.AuthHttpInterceptor(router))
	router.Handle("/{url:.*}", common.ErrorHandler(proxy.Serve)).Methods("GET")

	var listen = fmt.Sprintf("%s:%d", c.String("listen-ip"), c.Int("listen-port"))
	log.Infof("Listening on %s", listen)
	err := http.ListenAndServe(listen, nil)

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
