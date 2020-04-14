package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

		userData := make(map[string]string)

		inputMap := make(map[string]string)
		err := json.Unmarshal([]byte(bodyStr), &inputMap)
		if err != nil {
			sess.PushMsg("ERROR @ Unmarshal r.body, will try configs in cache, btw json.Unmarshal error = " + err.Error())
			userData = sess.UserData
			if sess.UserData["awsRegion"] == "" {
				sess.PushMsg("well, configs in cache is no good either, bye")
				return
			}
		} else {
			sess.PushMsg("config provided in r.body")
			userData = inputMap
		}

		awss, err := utils.NewAwsSvs(userData["awsKey"], userData["awsSecret"], userData["awsRegion"])
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

		// _, err = awss.CheckCERTarn(userData["cf_CERTarn"])
		// if err != nil {
		// 	sess.PushMsg("ERROR @ CheckCERTarn: " + err.Error())
		// 	return
		// }
		// sess.PushMsg("good CERTarn too")

		stackName := "iotcp-" + utils.StackNameGen()

		cfParams := []*cloudformation.Parameter{}

		for k := range userData {
			if k[0:3] == "cf_" {
				key := k[3:]
				val := userData[k]
				isPwdGen, _ := regexp.MatchString(`PwdGen\(\d+\)`, val)
				if isPwdGen {
					compRegEx := regexp.MustCompile(`PwdGen\((?P<len>\d+)\)`)
					lenStr := compRegEx.FindStringSubmatch(val)[1]
					userData[k] = val
					len, err := strconv.Atoi(lenStr)
					if err != nil {
						sess.PushMsg("ERROR @ parsing PwdGen length: " + err.Error())
						return
					}
					val = utils.PwdGen(len)
				}
				cfParams = append(cfParams,
					&cloudformation.Parameter{ParameterKey: aws.String(key), ParameterValue: aws.String(val)})
			}
		}

		go func() {
			err = awss.CreateCFstack(stackName, "https://"+userData["S3bucket"]+".s3.amazonaws.com/OneBtnDep.yaml", cfParams)
			if err != nil {
				sess.PushMsg("ERROR @ CreateCFstack for " + stackName + ": " + err.Error())
			}
			createSSMparam(stackName, accountNum, userData, awss, sess)
		}()
		sess.PushMsg("&#128640;CreateCFstack started for stackName=" + stackName)
		go reportCreateCFstackStatus(stackName, userData, sess, awss)
		return

	default:
		return
		// fmt.Fprintf(w, "unexpected method: "+r.Method)
	}
}

func createSSMparam(stackName string, accountNum string, userData map[string]string, awss *utils.AwsSvs, sess *utils.CacheBoxSessData) error {
	stacks, err := awss.GetStack(stackName)
	if err != nil {
		sess.PushMsg("ERROR @ createSSMparam -- GetStack: " + err.Error())
		return err
	}
	// userData := *sess.UserData
	//----------create SSM parameter
	paramMap, err := getSSMparamFromS3json(awss, userData, "ssmParam.json")
	if err != nil {
		sess.PushMsg("ERROR @ createSSMparamFromS3json: " + err.Error())
		return err
	}
	for _, k := range userData {
		if k[0:3] == "cf_" {
			paramMap[k[3:]] = userData[k]
		}
	}
	stackOutputs := stacks[0].Outputs
	for _, output := range stackOutputs {
		paramMap[*output.Description] = *output.OutputValue
	}

	paramJSONbytes, _ := json.Marshal(paramMap)
	err = awss.CreateSSMparameter(stackName, string(paramJSONbytes))
	if err != nil {
		sess.PushMsg("ERROR @ createSSMparamFromS3json: " + err.Error())
		return err
	}
	return nil
}

func getSSMparamFromS3json(awss *utils.AwsSvs, userData map[string]string, s3jsonFile string) (map[string]string, error) {
	err := awss.DownloadS3item(userData["S3bucket"], s3jsonFile)
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

func reportCreateCFstackStatus(stackName string, userData map[string]string, sess *utils.CacheBoxSessData, awss *utils.AwsSvs) error {
	time.Sleep(time.Second * 10)
	stackStatus := "something something IN_PROGRESS"
	for strings.Contains(stackStatus, "IN_PROGRESS") {
		stacks, err := awss.GetStack(stackName)
		if err != nil {
			sess.PushMsg("ERROR @ reportCreateCFstackStatus: " + err.Error())
			return err
		}
		stack := *stacks[0]
		stackStatus = *stack.StackStatus
		sinceStart := time.Now().UTC().Sub(stack.CreationTime.UTC()).Round(time.Second).String()
		stackLink := "https://" + userData["awsRegion"] + ".console.aws.amazon.com/cloudformation/home?region=" +
			userData["awsRegion"] + "#/stacks/stackinfo?stackId=" + *stack.StackId

		reportMsg := "<span style=\"color:white\">(" + sinceStart + ")</span> status of CF stack " +
			"<a href=\"" + stackLink + "\" target=\"_blank\"><b>&#128279;" + stackName + "</b></a>" + " is " + stackStatus
		if stack.StackStatusReason != nil {
			reportMsg = reportMsg + " because " + *stack.StackStatusReason
		}
		sess.PushMsg(reportMsg)
		time.Sleep(time.Second * 60)
	}
	return nil
}
