package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"text/template"

	"main/structs"
	"main/utils"
)

var RootHF = http.HandlerFunc(Root)

func Root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("./root.html")
	if err != nil {
		utils.Logger.Panic("failed to parse root.html template -- " + err.Error())
	}

	c, err := r.Cookie("name")
	if err != nil {
		t.Execute(w, nil)
		return
	}
	var userData structs.UserData
	err = json.Unmarshal([]byte(c.Value), &userData)

	awss, err := utils.NewAwsSvs(userData.AwsStuff.Key, userData.AwsStuff.Secret, userData.AwsStuff.Region)
	if err != nil {
		utils.Logger.Println(" (to be removed) bad cookie won't login: " + err.Error())
		c.MaxAge = -1
		http.SetCookie(w, c)
		t.Execute(w, nil)
		return
	} else if len(userData.Stacks) < 1 {
		utils.Logger.Println(" len(userData.Stacks) < 1")
		t.Execute(w, nil)
		return
	}

	utils.Logger.Println("------len(Stacks) = " + strconv.Itoa(len(userData.Stacks)))
	acctID, err := awss.GetAccountID()
	utils.Logger.Println("------acctID = " + acctID)

	for _, v := range userData.Stacks {
		stack, err := awss.GetStack(v.StackName)
		if err != nil || len(stack) < 1 {
			v.StackLink = "null"
			v.LastStatus = "null"
		}
		v.LastStatus = stack[0].LastUpdatedTime.String() + ": " + *stack[0].StackStatus + " -- " + *stack[0].StackStatusReason
		v.StackLink = "https://" + userData.AwsStuff.Region + ".console.aws.amazon.com/cloudformation/home?region=" +
			userData.AwsStuff.Region + "#/stacks/stackinfo?stackId=" + *stack[0].StackId
	}

	d := structs.RootPageData{
		ShowStackInfo: true,
		Stacks:        userData.Stacks,
	}

	t.Execute(w, d)
}

// var RootHF = http.HandlerFunc(
// 	func(w http.ResponseWriter, r *http.Request) {
// 		if r.URL.Path != "/" {
// 			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
// 			return
// 		}
// 		w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 		t, err := template.ParseFiles("./root.html")
// 		if err != nil {
// 			utils.Logger.Panic("failed to parse root.html template -- " + err.Error())
// 		}
// 		t.Execute(w, nil)
// 	})
