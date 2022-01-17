package main

import (
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type apiRequest struct {
	Group    string `json:"group"`
	KeyID    string `json:"key_id"`
	CallerID string `json:"caller_id"` // requester arn
}

type userOutput struct {
	User   *types.User
	Groups []types.Group
}
