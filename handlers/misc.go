package handlers

import "net/http"

var KeepAliveHF = http.HandlerFunc(KeepAlive)

func KeepAlive(w http.ResponseWriter, r *http.Request) {

}
