package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"maverics/app"
	"maverics/log"
	"maverics/session"
)

const (
	// The API used for the example returns data about users. The Service Extension
	// interacts with the '/users' endpoint.
	apiURL = "https://jsonplaceholder.typicode.com"

	// The API request made in this example filters the response by email address. To
	// search for a different user, update the email address below.
	exampleUserEmail = "Sincere@april.biz"
)

// user represents the user object returned from the API.
type user struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  struct {
		Street  string `json:"street"`
		Suite   string `json:"suite"`
		City    string `json:"city"`
		Zipcode string `json:"zipcode"`
	} `json:"address"`
}

// LoadAttrs is responsible for loading user attributes from a proprietary API. The
// loaded attributes are cached on the session so that they can be used as HTTP
// headers that are sent to the upstream application.
func LoadAttrs(ag *app.AppGateway, rw http.ResponseWriter, req *http.Request) error {
	log.Info("msg", "loading custom attribute from API")

	log.Info("msg", "building API request")
	// For this example, we use exampleUserEmail since the mock API only knows
	// about a static set of users. For production use, the user would normally be
	// looked up by doing something like session.GetString(req, "azure.email").
	filter := url.Values{}
	filter.Add("email", exampleUserEmail)

	apiReq, err := http.NewRequest(
		http.MethodGet,
		apiURL+"/users?"+filter.Encode(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create new HTTP request: %w", err)
	}

	log.Info("msg", "making API request to retrieve attributes")
	client := http.DefaultClient
	resp, err := client.Do(apiReq)
	if err != nil {
		return fmt.Errorf("failed to make API request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"recieved unexpected response code, expected HTTP 200 and receieved %d",
			resp.StatusCode,
		)
	}

	log.Info("msg", "decoding response from API")
	var users = make([]user, 0)
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if len(users) != 1 {
		return fmt.Errorf(
			"received unexpected response body, expected 1 user and received %d",
			len(users),
		)
	}
	log.Info("msg", "successfully loaded attributes from API, adding attributes to session")

	// Once attributes are stored on the session, they can be used for policy evaluation
	// and as HTTP headers.
	targetUser := users[0]
	session.Set(req, "api.id", targetUser.ID)
	session.Set(req, "api.name", targetUser.Name)
	session.Set(req, "api.username", targetUser.Username)
	session.Set(req, "api.email", targetUser.Email)
	session.Set(req, "api.phone", targetUser.Phone)
	session.Set(req, "api.address.street", targetUser.Address.Street)
	session.Set(req, "api.address.suite", targetUser.Address.Suite)
	session.Set(req, "api.address.city", targetUser.Address.City)
	session.Set(req, "api.address.zipcode", targetUser.Address.Zipcode)

	return nil
}
