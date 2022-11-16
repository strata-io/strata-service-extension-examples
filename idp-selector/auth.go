package main

import (
	"fmt"
	"net/http"

	"maverics/app"
	"maverics/log"
	"maverics/session"
)

// IsAuthenticated determines if the user is authenticated. Authentication status is
// derived by querying the session cache.
func IsAuthenticated(ag *app.AppGateway, _ http.ResponseWriter, req *http.Request) bool {
	log.Info("msg", "determining if user is authenticated")

	for idpName := range ag.IDPs {
		authenticated := session.GetString(req, idpName+".authenticated")
		if authenticated == "true" {
			log.Info("msg", fmt.Sprintf("user is authenticated with '%s'", idpName))
			return true
		}
	}

	return false
}

// Authenticate authenticates the user against the IDP that they select.
func Authenticate(ag *app.AppGateway, rw http.ResponseWriter, req *http.Request) error {
	log.Info("msg", "authenticating user")

	if req.Method == http.MethodGet {
		log.Info("msg", "received GET request, rendering IDP selector form")

		_, _ = rw.Write([]byte(idpForm))
		return nil
	}

	if req.Method != http.MethodPost {
		return fmt.Errorf("receieved unexpected request type '%s', expected POST", req.Method)
	}

	log.Info("msg", "parsing form from request")
	err := req.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form from request: %w", err)
	}

	selectedIDP := req.Form.Get("idp")
	log.Info("msg", fmt.Sprintf("user seleceted '%s' IDP for authentication", selectedIDP))

	log.Info("msg", fmt.Sprintf("authenticating user against '%s", selectedIDP))
	idp, found := ag.IDPs[selectedIDP]
	if !found {
		return fmt.Errorf("selected IDP '%s' was not found on AppGateway", selectedIDP)
	}
	idp.CreateRequest().Login(rw, req)
	return nil
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
