package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AwsSvs struct{ sess *session.Session }

func NewAwsSvs(key, secret, region string) (AwsSvs, error) {
	os.Setenv("AWS_ACCESS_KEY_ID", key)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secret)
	os.Setenv("AWS_REGION", region)
	sess, err := session.NewSession()
	if err != nil {
		return AwsSvs{}, err
	}
	return AwsSvs{sess: sess}, err
}

func (as AwsSvs) CreateCFstack(stackName string, templateURL string) error {
	svc := cloudformation.New(as.sess)

	input := &cloudformation.CreateStackInput{
		TemplateURL:  aws.String(templateURL),
		StackName:    aws.String(stackName),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
	}

	_, err := svc.CreateStack(input)
	if err != nil {
		return err
	}
	Logger.Println("======aws cloudformation stack creation started for stackName = " + stackName)

	desInput := &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}

	return svc.WaitUntilStackCreateComplete(desInput)
}

func (as AwsSvs) GetStackEvents(stackName string) ([]*cloudformation.StackEvent, error) {
	if stackName == "" {
		return nil, errors.New("missing <stackName>")
	}
	svc := cloudformation.New(as.sess)
	out, err := svc.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{StackName: aws.String(stackName)})
	if err != nil {
		return nil, err
	}
	return out.StackEvents, err
}
func (as AwsSvs) GetStack(stackName string) ([]*cloudformation.Stack, error) {
	if stackName == "" {
		return nil, errors.New("missing <stackName>")
	}
	svc := cloudformation.New(as.sess)
	out, err := svc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, err
	}
	return out.Stacks, err
}

func (as AwsSvs) CreateSSMparameter(paramName, paramValue string) error {

	svc := ssm.New(as.sess)

	req, _ := svc.PutParameterRequest(&ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Overwrite: aws.Bool(true),
		Type:      aws.String("SecureString"),
		Value:     aws.String(paramValue),
	})
	return req.Send()

}

func (as AwsSvs) DownloadS3item(bucket, item string) error {
	fmt.Println("---awsDownloadS3item---bucket =" + bucket + ", key=" + item)
	pos := strings.LastIndex(item, `/`)
	os.MkdirAll(item[:pos], os.ModePerm)
	file, err := os.Create(item)
	if err != nil {
		// Logger.Panic("---awsDownloadS3item --- failed to create file ---err = " + err.Error())
		return err
	}
	defer file.Close()
	// sess, _ := session.NewSession()

	downloader := s3manager.NewDownloader(as.sess)

	numBytes, err := downloader.Download(
		file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		// Logger.Panic("---awsDownloadS3item --- failed to download ---err = " + err.Error())
		return err
	}
	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
	return nil
}

func (as AwsSvs) GetAccountID() (string, error) {
	svc := sts.New(as.sess)
	r, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return *r.Account, nil
}
