package handlers

import (
	"main/utils"
	"net/http"
)

var EcsClusterDepHF = http.HandlerFunc(EcsClusterDep)

func EcsClusterDep(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/EcsClusterDep" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	sess := utils.GetSession(r.Cookie)
	sess.PushMsg("!!! EcsClusterDep started !!!")

	rData, err := utils.ParseJsonReqBody(r.Body)
	if err != nil {
		sess.PushMsg("ERROR @ Unmarshal r.body, will try configs in cache, btw json.Unmarshal error = " + err.Error())
		return
	}

	_ = rData
}
