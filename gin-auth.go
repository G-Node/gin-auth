package main

import (
	"fmt"
	"github.com/G-Node/gin-auth/data"
)

func main() {
	conf, err := data.LoadDbConf("conf/dbconf.yml")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(conf.Open)
	}

	err = data.InitDb(conf)
	if err != nil {
		fmt.Println(err)
	}
}
