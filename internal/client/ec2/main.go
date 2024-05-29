package ec2

const (
	serviceCode = "AmazonEC2"
)

type Client struct {
}

func New() *Client {
	return &Client{}
}
