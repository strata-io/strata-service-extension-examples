package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"maverics/app"
	"maverics/session"
)

// The following values should be updated for your environment.
const (
	// policyStoreID: The ID of your Amazon Verified Permissions policy store.
	policyStoreID = "your-verified-permissions-store-id"
	// awsKeyID: The key ID of an IAM user with read access to your Amazon Verified
	// Permissions policy store.
	awsKeyID = "your-aws-key-id"
	// awsSecretKey: The corresponding secret key for your IAM user.
	awsSecretKey = "your-aws-secret-key"
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
		"https://authz-verifiedpermissions.us-east-1.amazonaws.com/policy-stores/%s/is-authorized",
		policyStoreID,
	)

	avpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(postBody))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	credential := strings.Join([]string{
		now.UTC().Format("20060102"),
		"us-east-1",
		"verifiedpermissions",
		"aws4_request",
	}, "/")
	sig, err := signRequest(avpReq, now, credential)
	if err != nil {
		return nil, err
	}
	date := now.UTC().Format("20060102T150405Z")

	avpReq.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=host;x-amz-date, Signature=%s",
		awsKeyID,
		credential,
		sig,
	))
	avpReq.Header.Set("X-Amz-Date", date)

	return avpReq, nil
}

// signRequest generates a signature for the request to the is-authorized endpoint
// for the verified permission API.
func signRequest(req *http.Request, now time.Time, credential string) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	bodyHash := hex.EncodeToString(hashSHA256(body))
	req.Body = io.NopCloser(bytes.NewReader(body))

	payload := fmt.Sprintf(`POST
/policy-stores/%s/is-authorized

host:authz-verifiedpermissions.us-east-1.amazonaws.com
x-amz-date:%s

host;x-amz-date
%s`,
		policyStoreID,
		now.UTC().Format("20060102T150405Z"),
		bodyHash,
	)

	payloadHash := hex.EncodeToString(hashSHA256([]byte(payload)))

	sigPayload := fmt.Sprintf(`AWS4-HMAC-SHA256
%s
%s
%s`,
		now.UTC().Format("20060102T150405Z"),
		credential,
		payloadHash,
	)

	kDate := hmacSHA256([]byte("AWS4"+awsSecretKey), []byte(now.UTC().Format("20060102")))
	kRegion := hmacSHA256(kDate, []byte("us-east-1"))
	kService := hmacSHA256(kRegion, []byte("verifiedpermissions"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	signature := hmacSHA256(kSigning, []byte(sigPayload))

	return hex.EncodeToString(signature), nil
}

// hmacSHA256 is a helper method that generates a SHA256 signature of the provided
// data using the supplied key.
func hmacSHA256(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
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
