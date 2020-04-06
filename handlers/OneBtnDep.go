package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"main/structs"
	"main/utils"
)

var OneBtnDepHF = http.HandlerFunc(OneBtnDep)

func OneBtnDep(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/OneBtnDep" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	switch r.Method {
	// case "GET":
	// 	fmt.Println("oneBtn---received GET")
	case "POST":
		bodyStr := func() string { bytes, _ := ioutil.ReadAll(r.Body); return string(bytes) }()
		var userData structs.UserData

		err := json.Unmarshal([]byte(bodyStr), &userData.AwsStuff)
		if err != nil {
			utils.Logger.Println("---case---invalid config in r.body ... try cookie")
			c, err := r.Cookie("name")
			if err != nil {
				utils.Logger.Println("---failed--- no input and no cookie == bad ...... cookie err = " + err.Error())
				fmt.Fprintf(w, "ERROR -- missing config: "+err.Error())
				return
			}
			utils.Logger.Println("getting configs from cookie")
			json.Unmarshal([]byte(c.Value), &userData)
		} else {
			utils.Logger.Println("---case---config provided in r.body")
			// json.Unmarshal([]byte(bodyStr), &userData.AwsStuff)
		}
		awss, err := utils.NewAwsSvs(userData.AwsStuff.Key, userData.AwsStuff.Secret, userData.AwsStuff.Region)
		if err != nil {
			fmt.Fprintf(w, "ERROR @ NewAwsSvs: "+err.Error())
			return
		}
		accountNum, err := awss.GetAccountID()
		if err != nil {
			fmt.Fprintf(w, "ERROR @ GetAccountID: "+err.Error())
			return
		}
		stackName := "iotcp-" + utils.ShortMiniteUniqueID()
		userData.Stacks = append(userData.Stacks, structs.StackInfo{StackName: stackName, TimeStart: time.Now().UTC(), StackLink: "", LastStatus: ""})

		cookie := http.Cookie{Name: "name",
			Value: func() string { bytes, _ := json.Marshal(userData); return string(bytes) }(), Path: "/",
			// MaxAge: 3600, Secure: true, HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		err = awss.DownloadS3item(userData.AwsStuff.S3bucket, "1Btn/ssmParam.json")
		if err != nil {
			fmt.Fprintf(w, "ERROR @ DownloadS3item: "+err.Error())
			return
		}

		paramJSONbytes, _ := ioutil.ReadFile("1Btn/ssmParam.json")
		os.Remove("1Btn/ssmParam.json")
		paramMap := make(map[string]string)
		err = json.Unmarshal(paramJSONbytes, &paramMap)
		if err != nil {
			fmt.Fprintf(w, "ERROR @ json.Unmarshal(paramJSONbytes, &paramMap) = "+err.Error())
			return
		}
		paramMap["internalEndpoint"] = "http://nlb." + stackName
		paramMap["deviceNotificationQueue"] = "https://sqs." + userData.AwsStuff.Region + ".amazonaws.com/" + accountNum + "/" + stackName + "-cycleNotification"
		paramMap["deviceBillableEventQueue"] = "https://sqs." + userData.AwsStuff.Region + ".amazonaws.com/" + accountNum + "/" + stackName + "-billiableEvents"
		paramJSONbytes, _ = json.Marshal(paramMap)
		awss.CreateSSMparameter(stackName, string(paramJSONbytes))

		cfFileURL := "https://" + userData.AwsStuff.S3bucket + ".s3.amazonaws.com/1Btn/_iotcp.yaml"
		go awss.CreateCFstack(stackName, cfFileURL)

		fmt.Fprintf(w, "stackName="+stackName)
	default:
		fmt.Fprintf(w, "unexpected method: "+r.Method)
	}
}
