package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/web"
	"github.com/docopt/docopt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	major  = 0
	minor  = 1
	status = "Alpha"
)

func versionString() string {
	return fmt.Sprintf("gin-auth %d.%d %s", major, minor, status)
}

const doc = `G-Node Infrastructure Authentication Provider

Usage:
  gin-auth [--res <dir>]
  gin-auth -h | --help
  gin-auth --version

Options:
  --res <dir>     Path to the resources directory where configuration files,
                  templates and static files are located. By default gin-auth
                  will use GOPATH to find the directory.
  -h --help       Show this screen.
  --version       Print gin-auth version`

func main() {
	args, _ := docopt.Parse(doc, nil, true, versionString(), false)
	if res, ok := args["--res"]; ok && res != nil {
		conf.SetResourcesPath(res.(string))
	}

	srvConf := conf.GetServerConfig()
	conf.SmtpCheck()

	dbConf := conf.GetDbConfig()
	data.InitDb(dbConf)
	data.InitClients(conf.GetClientsConfigFile())

	router := mux.NewRouter()
	router.NotFoundHandler = &web.NotFoundHandler{}

	web.RegisterRoutes(router)

	handler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(router)
	handler = handlers.LoggingHandler(os.Stdout, handler)
	handler = handlers.CORS(
		handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Authorization"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "PUT", "POST", "DELETE"}),
	)(handler)

	data.RunCleaner()

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", srvConf.Host, srvConf.Port),
		Handler: handler,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
