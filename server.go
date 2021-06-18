package main

import (
	"log"
	"metamaskServer/api"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./conf")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("ReadInConfig fail:", err.Error())
		os.Exit(1)
	}

	addr := viper.GetString("rpcPort")
	lisp := viper.GetString("listenPort")
	chainId := viper.GetString("chainId")
	networkId := viper.GetString("networkId")
	ethTo := viper.GetString("ethTo")

	certf := viper.GetString("tls.cert")
	keyf := viper.GetString("tls.key")

	s := api.NewServer(addr, chainId, networkId, ethTo)
	http.HandleFunc("/", s.HandRequest)

	if certf != "" && keyf != "" {
		log.Println("Running Server...", lisp)
		err := http.ListenAndServeTLS(lisp, certf, keyf, nil)
		if err != nil {
			log.Println("start fasthttp fail:", err.Error())
			os.Exit(1)
		}
	}
	return
}
