package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func getClient() (*iam.Client, error) {
	role := os.Getenv("ASSUME_ROLE_ARN")
	if len(role) == 0 {
		return nil, errors.New("failed reading ASSUME_ROLE_ARN")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	creds := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), role)
	cfg.Credentials = aws.NewCredentialsCache(creds)

	return iam.NewFromConfig(cfg), nil
}

func getRequester(c *gin.Context) (apiRequest, string) {
	var req apiRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		handleError(c, http.StatusBadRequest, err.Error())
		return req, ""
	}

	if len(req.KeyID) == 0 || len(req.CallerID) == 0 {
		handleError(c, http.StatusBadRequest, "invalid request")
		return req, ""
	}

	split := strings.Split(req.CallerID, "/")
	if len(split) == 0 {
		handleError(c, http.StatusBadRequest, "failed parsing caller_id")
		return req, ""
	}
	requester := split[len(split)-1]

	return req, requester
}

func isAdmin(keyID, requester string) bool {
	var isAdmin bool
	adminGroup := os.Getenv("ADMIN_GROUP")

	if hasValidKey(keyID, requester) {
		g, err := getUserGroups(requester)
		if err != nil {
			log.Println(err.Error())
			return isAdmin
		}

		for _, v := range g.Groups {
			if adminGroup == *v.GroupName {
				isAdmin = true
				break
			}
		}
	}

	return isAdmin
}

func hasValidKey(key, username string) bool {
	var isValidKey bool
	svc, err := getClient()
	if err != nil {
		log.Println(err.Error())
		return isValidKey
	}

	i := iam.ListAccessKeysInput{
		UserName: aws.String(username),
	}

	keys, err := svc.ListAccessKeys(context.TODO(), &i)
	if err != nil {
		log.Println(err.Error())
		return isValidKey
	}

	for _, v := range keys.AccessKeyMetadata {
		if key == *v.AccessKeyId {
			if v.Status == types.StatusTypeActive {
				isValidKey = true
			}
		}
	}

	return isValidKey
}
