package toolbag

import "time"

type UserData struct {
	AwsStuff AwsStuff    `json:"aws_stuff"`
	Stacks   []StackInfo `json:"stacks"`
}

type AwsStuff struct {
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	Region   string `json:"region"`
	S3bucket string `json:"s3bucket"`
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
