package main

import (
	"maverics/app"
	"maverics/log"
	"maverics/session"

	"net/http"
)

// CreateFirstNameHeader creates a custom first name header. The user's first name
// will be concatenated with "the Great".
func CreateFirstNameHeader(
	ag *app.AppGateway,
	rw http.ResponseWriter,
	req *http.Request,
) (http.Header, error) {
	log.Debug("msg", "retrieving first name from session")
	name := session.GetString(req, "azure.given_name") + " the Great"

	log.Debug("msg", "building custom first name header")
	header := make(http.Header)
	header["EXAMPLE-FIRST-NAME"] = []string{name}
	return header, nil
}

// CreateLastNameHeader creates a custom last name header. The user's last name
// will be prepended with "Dr.".
func CreateLastNameHeader(
	ag *app.AppGateway,
	rw http.ResponseWriter,
	req *http.Request,
) (http.Header, error) {
	log.Debug("msg", "retrieving last name from session")
	lastName := "Dr. " + session.GetString(req, "azure.family_name")

	log.Debug("msg", "building custom last name header")
	header := make(http.Header)
	header["EXAMPLE-LAST-NAME"] = []string{lastName}
	return header, nil
}
