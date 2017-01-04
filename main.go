package main

import (
	"fmt"
	"net/http"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/util"
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
  gin-auth [--res <dir>] [--conf <dir>]
  gin-auth -h | --help
  gin-auth --version

Options:
  --res <dir>     Path to the resources directory where templates
                  and static files are located. By default gin-auth
                  will use GOPATH to find the directory.
  --conf <dir>    Path to the configuration files directory. By default
                  gin-auth will use the resources/conf directory.
  -h --help       Show this screen.
  --version       Print gin-auth version`

func main() {
	args, _ := docopt.Parse(doc, nil, true, versionString(), false)
	if res, ok := args["--res"]; ok && res != nil {
		conf.SetResourcesPath(res.(string))
	}

	if config, ok := args["--conf"]; ok && config != nil {
		conf.SetConfigPath(config.(string))
	}

	// Initialize logging and make sure log files will be closed.
	logEnv := conf.GetLogEnv()
	defer logEnv.Close()

	srvConf := conf.GetServerConfig()
	err := conf.SmtpCheck()
	if err != nil {
		panic(err.Error())
	}

	dbConf := conf.GetDbConfig()
	data.InitDb(dbConf)
	data.InitClients(conf.GetClientsConfigFile())

	// Initialize externals
	conf.GetExternals()

	router := mux.NewRouter()
	router.NotFoundHandler = &web.NotFoundHandler{}

	web.RegisterRoutes(router)

	handler := util.RecoveryHandler(router, logEnv.Err, true)
	handler = handlers.LoggingHandler(logEnv.Access.Out, handler)
	handler = handlers.CORS(
		handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Authorization"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "PUT", "POST", "DELETE"}),
	)(handler)

	data.RunCleaner()
	data.RunEmailDispatch()

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", srvConf.Host, srvConf.Port),
		Handler: handler,
	}
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
