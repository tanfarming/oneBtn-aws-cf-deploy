package main

import (
	"main/handlers"
	"main/toolbag"
	"net/http"
)

func main() {

	// pwd, _ := os.Getwd()

	router := http.NewServeMux()

	router.Handle("/_files/", http.StripPrefix("/_files/", http.FileServer(http.Dir("_files"))))

	router.Handle("/", handlers.Root)
	router.Handle("/oneBtn", handlers.OneBtnDep)

	toolbag.StartServer(router, 888)
}
