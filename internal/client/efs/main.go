package efs

const (
	serviceCode = "AmazonEFS"
)

type awsEFSClient struct {
}

func New() *awsEFSClient {
	return &awsEFSClient{}
}
