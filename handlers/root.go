package handlers

import (
	"net/http"
	"text/template"

	"main/utils"
)

var RootHF = http.HandlerFunc(Root)

func Root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("./_templates/root.html")
	if err != nil {
		utils.Logger.Panic("failed to parse root.html template -- " + err.Error())
	}

	c, err := r.Cookie(utils.SessionTokenName)
	if err != nil {
		newCookie := utils.CreateNewSession()
		http.SetCookie(w, newCookie)
		t.Execute(w, nil)
		return
	}

	sess := utils.CACHE.Load(c.Value)
	if sess == nil {
		http.SetCookie(w, utils.CreateNewSession())
		t.Execute(w, nil)
		return
	}

	t.Execute(w, nil)

	// awss, err := utils.NewAwsSvs(sess.UserData.AwsStuff.Key, sess.UserData.AwsStuff.Secret, sess.UserData.AwsStuff.Region)
	// acctID, err := awss.GetAccountID()
	// if err != nil {
	// 	sess.PushMsg("bad aws configs: " + err.Error())
	// 	t.Execute(w, nil)
	// 	return
	// }
	// sess.PushMsg("good aws configs -- acctID = " + acctID)

	// if len(sess.UserData.Stacks) < 1 {
	// 	sess.PushMsg("but len(userData.Stacks) < 1 ")
	// 	t.Execute(w, nil)
	// 	return
	// }

	// sess.PushMsg("len(Stacks) = " + strconv.Itoa(len(sess.UserData.Stacks)))

	// for _, v := range sess.UserData.Stacks {
	// 	stack, err := awss.GetStack(v.StackName)
	// 	if err != nil || len(stack) < 1 {
	// 		v.StackLink = "null"
	// 		v.LastStatus = "null"
	// 	}
	// 	v.LastStatus = stack[0].LastUpdatedTime.String() + ": " + *stack[0].StackStatus + " -- " + *stack[0].StackStatusReason
	// 	v.StackLink = "https://" + sess.UserData.AwsStuff.Region + ".console.aws.amazon.com/cloudformation/home?region=" +
	// 		sess.UserData.AwsStuff.Region + "#/stacks/stackinfo?stackId=" + *stack[0].StackId
	// }

	// d := utils.RootPageData{
	// 	ShowStackInfo: true,
	// 	Stacks:        sess.UserData.Stacks,
	// }

	// t.Execute(w, d)
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
