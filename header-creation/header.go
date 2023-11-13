package main

import (
	"fmt"
	"net/http"

	"github.com/strata-io/service-extension/orchestrator"
)

// CreateFirstNameHeader creates a custom first name header. The user's first name
// will be concatenated with "the Great".
func CreateFirstNameHeader(
	api orchestrator.Orchestrator,
	_ http.ResponseWriter,
	_ *http.Request,
) (http.Header, error) {
	logger := api.Logger()
	logger.Info("se", "building first name custom header")

	session, err := api.Session()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve session: %w", err)
	}
	logger.Debug("se", "retrieving first name from session")
	firstName, err := session.GetString("azure.given_name")
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve attribute 'azure.given_name': %w", err)
	}
	name := firstName + " the Great"

	header := make(http.Header)
	header["EXAMPLE-FIRST-NAME"] = []string{name}
	return header, nil
}

// CreateLastNameHeader creates a custom last name header. The user's last name
// will be prepended with "Dr.".
func CreateLastNameHeader(
	api orchestrator.Orchestrator,
	_ http.ResponseWriter,
	_ *http.Request,
) (http.Header, error) {
	logger := api.Logger()
	logger.Info("se", "building last name custom header")

	session, err := api.Session()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve session: %w", err)
	}
	lastName, err := session.GetString("azure.family_name")
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve attribute 'azure.family_name': %w", err)
	}

	nameWithTitle := "Dr. " + lastName
	header := make(http.Header)
	header["EXAMPLE-LAST-NAME"] = []string{nameWithTitle}
	return header, nil
}
