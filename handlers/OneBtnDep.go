package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"main/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var OneBtnDepHF = http.HandlerFunc(OneBtnDep)

func OneBtnDep(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/OneBtnDep" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	cookie, _ := r.Cookie(utils.SessionTokenName)
	sess := utils.CACHE.Load(cookie.Value)
	sess.PushMsg("yep you just clicked THE ONE BUTTON")

	switch r.Method {
	// case "GET":
	// 	fmt.Println("oneBtn---received GET")
	case "POST":
		bodyStr := func() string { bytes, _ := ioutil.ReadAll(r.Body); return string(bytes) }()
		var userData utils.UserData

		err := json.Unmarshal([]byte(bodyStr), &userData.AwsStuff)
		if err != nil {
			sess.PushMsg("valid config NOT found in r.body == will try configs in cache ... btw json.Unmarshal error = " + err.Error())
			userData = *sess.UserData
			if userData.AwsStuff.Region == "" {
				sess.PushMsg("well, configs in cache is no good either, bye")
				return
			}
		} else {
			sess.PushMsg("config provided in r.body")
			sess.UserData = &userData
		}
		// t := reflect.TypeOf(userData.AwsStuff)
		// for i := 0; i < t.NumField(); i++ {
		// 	fmt.Printf("%+v\n", t.Field(i))
		// }
		// v := reflect.ValueOf(userData.AwsStuff)
		// for i := 0; i < v.NumField(); i++ {
		// 	fmt.Println(v.Field(i))
		// }
		// fmt.Println(userData.AwsStuff)
		if userData.AwsStuff.S3bucket == "" {
			userData.AwsStuff.S3bucket = "4thiq-onebtndep-oh"
		}
		if userData.AwsStuff.LambdaBucket == "" {
			userData.AwsStuff.LambdaBucket = "4thiq-lambdas-oh"
		}
		if userData.AwsStuff.ContainerRegistry == "" {
			userData.AwsStuff.ContainerRegistry = "614375816418.dkr.ecr.us-east-1.amazonaws.com"
		}
		//=======================================================================

		awss, err := utils.NewAwsSvs(userData.AwsStuff.Key, userData.AwsStuff.Secret, userData.AwsStuff.Region)
		if err != nil {
			sess.PushMsg("ERROR @ NewAwsSvs: " + err.Error())
			return
		}
		accountNum, err := awss.GetAccountID()
		if err != nil {
			sess.PushMsg("ERROR @ GetAccountID: " + err.Error())
			return
		}
		sess.PushMsg("good aws config for account #: " + accountNum)

		_, err = awss.CheckCERTarn(userData.AwsStuff.CERTarn)
		if err != nil {
			sess.PushMsg("ERROR @ CheckCERTarn: " + err.Error())
			return
		}
		sess.PushMsg("good CERTarn too")

		stackName := "iotcp-" + utils.StackNameGen()
		userData.Stacks = append(userData.Stacks,
			utils.StackInfo{
				StackName: stackName,
				TimeStart: time.Now().UTC()})

		dbPwd := utils.PwdGen(17)
		awsDislikedChars := []string{"/", "@", "\"", " "}
		for _, c := range awsDislikedChars {
			dbPwd = strings.Replace(dbPwd, c, "8", -1)
		}

		cfParams := []*cloudformation.Parameter{
			&cloudformation.Parameter{ParameterKey: aws.String("CERTarn"), ParameterValue: aws.String(userData.AwsStuff.CERTarn)},
			&cloudformation.Parameter{ParameterKey: aws.String("lambdaBucket"), ParameterValue: aws.String(userData.AwsStuff.LambdaBucket)},
			&cloudformation.Parameter{ParameterKey: aws.String("IMGreg"), ParameterValue: aws.String(userData.AwsStuff.ContainerRegistry)},
			&cloudformation.Parameter{ParameterKey: aws.String("dbPwd"), ParameterValue: aws.String(dbPwd)},
		}

		go func() {
			err = awss.CreateCFstack(stackName, "https://"+userData.AwsStuff.S3bucket+".s3.amazonaws.com/OneBtnDep.yaml", cfParams)
			if err != nil {
				sess.PushMsg("ERROR @ CreateCFstack for " + stackName + ": " + err.Error())
			}
			updateSSMparam(stackName, accountNum, userData, awss, sess)
		}()
		go reportCreateCFstackStatus(stackName, sess, awss)

		sess.PushMsg("CreateCFstack started for stackName=" + stackName)
		return
	default:
		fmt.Fprintf(w, "unexpected method: "+r.Method)
	}
}

func updateSSMparam(stackName string, accountNum string, userData utils.UserData, awss *utils.AwsSvs, sess *utils.CacheBoxSessData) {
	stacks, err := awss.GetStack(stackName)
	if err != nil {
		utils.Logger.Panic("ERROR @ GetStack: " + err.Error())
	}
	// userData := *sess.UserData
	//----------create SSM parameter
	paramMap, err := getSSMparamFromS3json(awss, userData, "ssmParam.json")
	if err != nil {
		sess.PushMsg("ERROR @ createSSMparamFromS3json: " + err.Error())
		return
	}
	paramMap["internalEndpoint"] = "http://nlb." + stackName
	paramMap["deviceNotificationQueue"] = "https://sqs." + userData.AwsStuff.Region + ".amazonaws.com/" + accountNum + "/" + stackName + "-cycleNotification"
	paramMap["deviceBillableEventQueue"] = "https://sqs." + userData.AwsStuff.Region + ".amazonaws.com/" + accountNum + "/" + stackName + "-billiableEvents"

	stackOutputs := stacks[0].Outputs
	for _, output := range stackOutputs {
		paramMap[*output.Description] = *output.OutputValue
	}

	paramJSONbytes, _ := json.Marshal(paramMap)
	err = awss.CreateSSMparameter(stackName, string(paramJSONbytes))
	if err != nil {
		sess.PushMsg("ERROR @ createSSMparamFromS3json: " + err.Error())
		return
	}
}

func getSSMparamFromS3json(awss *utils.AwsSvs, userData utils.UserData, s3jsonFile string) (map[string]string, error) {
	err := awss.DownloadS3item(userData.AwsStuff.S3bucket, s3jsonFile)
	if err != nil {
		return nil, err
	}

	paramJSONbytes, _ := ioutil.ReadFile(s3jsonFile)
	os.Remove(s3jsonFile)
	paramMap := make(map[string]string)
	err = json.Unmarshal(paramJSONbytes, &paramMap)
	if err != nil {
		return nil, err
	}
	return paramMap, nil
}

func reportCreateCFstackStatus(stackName string, sess *utils.CacheBoxSessData, awss *utils.AwsSvs) {
	time.Sleep(time.Second * 15)
	stackStatus := "something something IN_PROGRESS"
	for strings.Contains(stackStatus, "IN_PROGRESS") {
		stacks, err := awss.GetStack(stackName)
		if err != nil {
			utils.Logger.Panic("ERROR @ reportCreateCFstackStatus: " + err.Error())
		}
		stack := *stacks[0]
		stackStatus = *stack.StackStatus
		sinceStart := time.Now().UTC().Sub(stack.CreationTime.UTC()).Round(time.Second).String()
		stackLink := "https://" + sess.UserData.AwsStuff.Region + ".console.aws.amazon.com/cloudformation/home?region=" +
			sess.UserData.AwsStuff.Region + "#/stacks/stackinfo?stackId=" + *stack.StackId
		reportMsg := "<span style=\"color:white\">(" + sinceStart + ")</span> status of CF stack " +
			"<a href=\"" + stackLink + "\">&#128279;<b>" + stackName + "</b></a>" + " is " + stackStatus
		if stack.StackStatusReason != nil {
			reportMsg = reportMsg + " because " + *stack.StackStatusReason
		}
		sess.PushMsg(reportMsg)
		time.Sleep(time.Second * 30)
	}
	return
}
