package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"main/toolbag"
)

var logger = toolbag.Logger

var Root = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("./root.html")
	if err != nil {
		logger.Panic("failed to parse root.html template -- " + err.Error())
	}

	c, err := r.Cookie("name")
	if err != nil {
		t.Execute(w, nil)
		return
	}
	var userData toolbag.UserData
	err = json.Unmarshal([]byte(c.Value), &userData)

	awss, err := toolbag.NewAwsSvs(userData.AwsStuff.Key, userData.AwsStuff.Secret, userData.AwsStuff.Region)
	if err != nil {
		fmt.Println("root---got cookie but ERROR login aws @ NewAwsSvs: " + err.Error())
		t.Execute(w, nil)
		return
	}

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

	t.Execute(w, nil)
})
