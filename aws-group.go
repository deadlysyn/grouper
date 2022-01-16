package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func getGroups() (*iam.ListGroupsOutput, error) {
	svc, err := getClient()
	if err != nil {
		return nil, err
	}

	gi := iam.ListGroupsInput{}
	g, err := svc.ListGroups(context.TODO(), &gi)
	if err != nil {
		return nil, err
	}

	if g.IsTruncated {
		gi.Marker = g.Marker
		for {
			gg, err := svc.ListGroups(context.TODO(), &gi)
			if err != nil {
				return nil, err
			}
			g.Groups = append(g.Groups, gg.Groups...)
			gi.Marker = gg.Marker
			if !gg.IsTruncated {
				break
			}
		}
	}

	return g, nil
}

func getGroupUsers(groupname string) (*iam.GetGroupOutput, error) {
	svc, err := getClient()
	if err != nil {
		return nil, err
	}

	gi := iam.GetGroupInput{
		GroupName: aws.String(groupname),
	}
	g, err := svc.GetGroup(context.TODO(), &gi)
	if err != nil {
		return nil, err
	}

	if g.IsTruncated {
		gi.Marker = g.Marker
		for {
			gg, err := svc.GetGroup(context.TODO(), &gi)
			if err != nil {
				return nil, err
			}
			g.Users = append(g.Users, gg.Users...)
			gi.Marker = gg.Marker
			if !gg.IsTruncated {
				break
			}
		}
	}

	return g, nil
}

func getUserGroups(username string) (*iam.ListGroupsForUserOutput, error) {
	svc, err := getClient()
	if err != nil {
		return nil, err
	}

	gi := iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	}
	g, err := svc.ListGroupsForUser(context.TODO(), &gi)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func updateGroup(group, member, requester string) error {
	svc, err := getClient()
	if err != nil {
		return err
	}

	g, err := getUserGroups(requester)
	if err != nil {
		return err
	}

	var groupMember bool
	for _, v := range g.Groups {
		// "admin" group can manage all groups
		if group == *v.GroupName || "admin" == *v.GroupName {
			groupMember = true
		}
	}

	if !groupMember {
		return fmt.Errorf("%s is not a member of %s", requester, group)
	}

	i := iam.AddUserToGroupInput{
		GroupName: aws.String(group),
		UserName:  aws.String(member),
	}
	_, err = svc.AddUserToGroup(context.TODO(), &i)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("%s added %s to AWS IAM group %s", requester, member, group)
	err = slackNotify(msg)
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}
