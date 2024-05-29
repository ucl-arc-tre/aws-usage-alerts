package types

type EmailAddress string

type Group string

type ResourceUsage map[Group]Cost

type AWSUsage struct {
	EFS ResourceUsage
	EC2 ResourceUsage
}
