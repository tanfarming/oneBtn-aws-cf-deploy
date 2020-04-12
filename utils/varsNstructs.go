package utils

import (
	"log"
	"os"
	"time"
)

const SessionTokenName = "session_token"

var Logger = log.New(os.Stdout, "http: ", log.LstdFlags)
var CACHE = NewCacheBox()

type RootPageData struct {
	ShowStackInfo bool
	Stacks        []StackInfo
}

type UserData struct {
	AwsStuff AwsStuff    `json:"aws_stuff"`
	Stacks   []StackInfo `json:"stacks"`
}

type AwsStuff struct {
	Key               string `json:"key"`
	Secret            string `json:"secret"`
	Region            string `json:"region"`
	CERTarn           string `json:"CERTarn"`
	S3bucket          string `json:"S3bucket,omitempty"`
	LambdaBucket      string `json:"LambdaBucket,omitempty"`
	ContainerRegistry string `json:"ContainerRegistry,omitempty"`
}

type StackInfo struct {
	StackName  string    `json:"stack_name"`
	TimeStart  time.Time `json:"time_start"`
	StackLink  string
	LastStatus string
}

func newUserData() UserData {
	userData := UserData{
		AwsStuff: AwsStuff{},
		Stacks:   []StackInfo{StackInfo{}}}
	return userData
}

type stackInfoData struct {
	showStackInfo bool
	Stacks        []StackInfo
}
