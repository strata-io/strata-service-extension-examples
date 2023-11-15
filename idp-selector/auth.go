package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/strata-io/service-extension/orchestrator"
)

// IsAuthenticated determines if the user is authenticated. Authentication status is
// derived by querying the session cache.
func IsAuthenticated(api orchestrator.Orchestrator, rw http.ResponseWriter, _ *http.Request) bool {
	logger := api.Logger()
	logger.Info("se", "determining if user is authenticated")

	session, err := api.Session()
	if err != nil {
		logger.Error("se", "unable to retrieve session", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return false
	}

	metadata := api.Metadata()
	idpNames := strings.Split(metadata["idps"].(string), ",")

	for _, idpName := range idpNames {
		authenticated, err := session.GetString(idpName + ".authenticated")
		if err != nil {
			logger.Error(
				"se", fmt.Sprintf("unable to retrieve session value '%s.authenticated'", idpName),
				"error", err.Error(),
			)
		}
		if authenticated == "true" {
			logger.Info("se", fmt.Sprintf("user is authenticated with '%s'", idpName))
			return true
		}
	}

	return false
}

// Authenticate authenticates the user against the IDP that they select.
func Authenticate(api orchestrator.Orchestrator, rw http.ResponseWriter, req *http.Request) {
	logger := api.Logger()
	logger.Info("se", "authenticating user")

	if req.Method == http.MethodGet {
		logger.Info("se", "received GET request, rendering IDP selector form")
		_, _ = rw.Write([]byte(idpForm))
		return
	}

	if req.Method != http.MethodPost {
		logger.Error("se", fmt.Sprintf("received unexpected request method '%s', expected POST", req.Method))
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	logger.Debug("se", "parsing form from request")
	err := req.ParseForm()
	if err != nil {
		logger.Error("se", "failed to parse form from request", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	selectedIDP := req.Form.Get("idp")
	logger.Info("se", fmt.Sprintf("user selected '%s' IDP for authentication", selectedIDP))

	idp, err := api.IdentityProvider(selectedIDP)
	if err != nil {
		logger.Error("se", "unable to lookup idp", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	idp.Login(rw, req)
	return
}

// idpForm is a basic form that is rendered in order to enable the user to pick which
// IDP they want to authenticate against. The markup can be styled as necessary,
// loaded from an external file, be rendered as a dynamic template, etc.
const idpForm = `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>IDP Selector</title>
    <meta charset="utf-8">
</head>

<body>
<div>
    <form method="post">
        <label for="idp">Select IDP:</label>
        <select id="idp" name="idp">
            <option value="azure">Azure</option>
            <option value="auth0">Auth0</option>
        </select>
        <input type="submit">
    </form>
</div>
</body>
</html>
	`
