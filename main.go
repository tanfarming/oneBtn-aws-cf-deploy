package main

import (
	"net/http"

	"main/handlers"
	"main/utils"
)

func main() {

	// pwd, _ := os.Getwd()
	router := http.NewServeMux()

	router.Handle("/_files/", http.StripPrefix("/_files/", http.FileServer(http.Dir("_files"))))

	router.Handle("/", handlers.RootHF)
	router.Handle("/OneBtnDep", handlers.OneBtnDepHF)
	router.Handle("/LogStream", handlers.LogStreamHF)

	router.Handle("/KeepAlive", handlers.KeepAliveHF)

	utils.StartServer(router, 888)
}
