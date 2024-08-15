package types

import "fmt"

type ResourceUsage map[Group]Cost

func (r ResourceUsage) Errors() []error {
	errors := []error{}
	for group, cost := range r {
		for _, err := range cost.Errors {
			if err != nil {
				errors = append(errors, fmt.Errorf("%v %v", group, err))
			}
		}
	}
	return errors
}

type AWSUsage struct {
	EFS ResourceUsage
	EC2 ResourceUsage
}

func (u *AWSUsage) Errors() []error {
	errors := []error{}
	for _, err := range u.EFS.Errors() {
		errors = append(errors, fmt.Errorf("EFS:%v", err))
	}
	for _, err := range u.EC2.Errors() {
		errors = append(errors, fmt.Errorf("EC2:%v", err))
	}
	return errors
}
