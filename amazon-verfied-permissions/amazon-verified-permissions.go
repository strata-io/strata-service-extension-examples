package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"maverics/app"
	"maverics/session"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

// The following values should be updated for your environment.
const (
	// policyStoreID: The ID of your Amazon Verified Permissions policy store.
	policyStoreID = "your-verified-permissions-store-id"
	// The region of your Amazon Verified Permissions policy store.
	region = "your-region"
	// The session value from your IdP used as the principal ID in the call to
	// Amazon Verified Permissions.
	principalID = "Amazon_Cognito.email"
)

// IsAuthorized is called after you log in with your IdP.  This function calls the
// Amazon Verified Permissions API with the associated principalID and endpoint to
// determine the authorization decision.
func IsAuthorized(ag *app.AppGateway, rw http.ResponseWriter, req *http.Request) bool {
	email := session.GetString(req, principalID)
	log.Println("requesting isAuthorized decision for " + email + " at " + req.URL.Path)

	avpReq, err := createVerifiedPermissionsRequest(email, req.URL.Path)
	if err != nil {
		log.Println("error creating request: " + err.Error())
		return false
	}

	result, err := http.DefaultClient.Do(avpReq)
	if err != nil {
		log.Println("error sending request: " + err.Error())
		return false
	}
	responseBody, err := io.ReadAll(result.Body)
	if err != nil {
		log.Println("error reading response: " + err.Error())
		return false
	}

	var response Response
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Println("error unmarshalling response: " + err.Error())
		return false
	}

	log.Println("isAuthorized decision from Amazon verified permissions: " + string(responseBody))
	return response.Decision == "Allow"
}

// createVerifiedPermissionsRequest builds a new verified permissions API request with the supplied
// principal and path.
func createVerifiedPermissionsRequest(principal, path string) (*http.Request, error) {
	reqBody := Request{
		Action: Action{
			ActionId:   "view",
			ActionType: "Action",
		},
		Principal: Principal{
			EntityId:   principal,
			EntityType: "User",
		},
		Resource: Resource{
			EntityId:   path,
			EntityType: "Endpoint",
		},
	}
	postBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"https://authz-verifiedpermissions.%s.amazonaws.com/policy-stores/%s/is-authorized",
		region, policyStoreID,
	)

	avpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(postBody))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	payloadHash := hex.EncodeToString(hashSHA256(postBody))

	signer := v4.NewSigner()
	err = signer.SignHTTP(ctx, creds, avpReq, payloadHash, "verifiedpermissions", region, now)
	if err != nil {
		return nil, err
	}

	return avpReq, nil
}

// hashSHA256 is a helper method that returns a SHA256 hash of the provided data.
func hashSHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

type Request struct {
	Action    Action    `json:"Action"`
	Principal Principal `json:"Principal"`
	Resource  Resource  `json:"Resource"`
}

type Action struct {
	ActionId   string `json:"ActionId"`
	ActionType string `json:"ActionType"`
}
type Principal struct {
	EntityId   string `json:"EntityId"`
	EntityType string `json:"EntityType"`
}

type Resource struct {
	EntityId   string `json:"EntityId"`
	EntityType string `json:"EntityType"`
}

type Response struct {
	Decision            string        `json:"Decision"`
	DeterminingPolicies []interface{} `json:"DeterminingPolicies"`
	Errors              []interface{} `json:"Errors"`
}
