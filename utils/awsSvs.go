package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AwsSvs struct{ sess *session.Session }

func NewAwsSvs(key, secret, region string) (*AwsSvs, error) {
	os.Setenv("AWS_ACCESS_KEY_ID", key)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secret)
	os.Setenv("AWS_DEFAULT_REGION", region)
	os.Setenv("AWS_REGION", region)
	sess, err := session.NewSession()
	// sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})

	if err != nil {
		return &AwsSvs{}, err
	}
	return &AwsSvs{sess: sess}, err
}

func BuildCFlink(region, stackID string) string {
	return "https://" + region + ".console.aws.amazon.com/cloudformation/home?region=" + region + "#/stacks/stackinfo?stackId=" + stackID
}

func (as AwsSvs) CreateCFstack(stackName string, templateURL string, params []*cloudformation.Parameter) error {
	svc := cloudformation.New(as.sess)

	input := &cloudformation.CreateStackInput{
		TemplateURL:  aws.String(templateURL),
		StackName:    aws.String(stackName),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
		Parameters:   params,
	}
	_, err := svc.CreateStack(input)
	if err != nil {
		// Logger.Println("[ERROR]: CreateCFstack FAILED and error = " + err.Error())
		return err
	}
	// Logger.Println("===CreateCFstack started for stackName = " + stackName)

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
	if pos > 0 {
		os.MkdirAll(item[:pos], os.ModePerm)
	}

	file, err := os.Create(item)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(as.sess)

	numBytes, err := downloader.Download(
		file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
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

func (as AwsSvs) CheckCERTarn(CERTarn string) (cert string, err error) {
	svc := acm.New(as.sess)
	certOut, err := svc.GetCertificate(&acm.GetCertificateInput{CertificateArn: aws.String(CERTarn)})
	if err != nil {
		return "", err
	}
	return *certOut.Certificate, err
}
