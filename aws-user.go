package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func getUser(username string) (userOutput, error) {
	var out userOutput

	svc, err := getClient()
	if err != nil {
		return out, err
	}

	ui := iam.GetUserInput{
		UserName: aws.String(username),
	}
	u, err := svc.GetUser(context.TODO(), &ui)
	if err != nil {
		return out, err
	}

	g, err := getUserGroups(username)
	if err != nil {
		return out, err
	}

	// iam.GetUserOutput{}
	// iam.ListGroupsForUserOutput{}

	// out.User = u
	// out.Groups = g

	out.User = u.User
	out.Groups = g.Groups

	return out, nil
}
