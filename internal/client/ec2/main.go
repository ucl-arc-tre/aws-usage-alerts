package ec2

const (
	serviceCode = "AmazonEC2"
)

type awsEC2Client struct {
}

func New() *awsEC2Client {
	return &awsEC2Client{}
}
