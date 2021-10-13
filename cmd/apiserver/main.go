package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	"github.com/katelinlis/AuthBackend/internal/app/apiserver"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "/configs/apiserver-prod.toml", "path to config file")
}

func main() {
	flag.Parse()
	config := apiserver.NewConfig()

	path, err := os.Getwd()
	fmt.Println(path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = toml.DecodeFile(path+configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	go apiserver.Start(config)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
