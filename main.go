package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/web"
)

func main() {
	conf, err := data.LoadDbConf("conf/dbconf.yml")
	if err != nil {
		panic(err)
	}
	err = data.InitDb(conf)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.NotFoundHandler = &web.Error404{}

	web.RegisterRoutes(router)

	handler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(router)
	handler = handlers.LoggingHandler(os.Stdout, handler)

	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	server.ListenAndServe()
}
