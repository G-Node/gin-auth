package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"fmt"
	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/data"
	"github.com/G-Node/gin-auth/web"
)

func main() {
	srvConf := conf.GetServerConfig()
	dbConf := conf.GetDbConfig()

	data.InitDb(dbConf)

	router := mux.NewRouter()
	router.NotFoundHandler = &web.NotFoundHandler{}
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("resources/static/"))))

	web.RegisterRoutes(router)

	handler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(router)
	handler = handlers.LoggingHandler(os.Stdout, handler)

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", srvConf.Host, srvConf.Port),
		Handler: handler,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
