package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	ldap3 "github.com/go-ldap/ldap/v3"
	"github.com/strata-io/service-extension/orchestrator"
)

// IsAuthenticated determines if the user has been authenticated.
func IsAuthenticated(api orchestrator.Orchestrator, _ http.ResponseWriter, _ *http.Request) bool {
	logger := api.Logger()
	session, err := api.Session()
	if err != nil {
		logger.Error("se", "unable to retrieve session", "error", err.Error())
		return false
	}
	metadata := api.Metadata()
	idpName := metadata["idpName"].(string)

	logger.Debug("se", "determining if user is authenticated")

	isAuthenticated, err := session.GetString(fmt.Sprintf("%s.authenticated", idpName))
	if err != nil {
		logger.Error(
			"se", fmt.Sprintf("unable to retrieve session value '%s.authenticated'", idpName),
			"error", err.Error(),
		)
		return false
	}
	if isAuthenticated == "true" {
		return true
	}

	return false
}

// Authenticate authenticates the user against the IDP.
func Authenticate(api orchestrator.Orchestrator, rw http.ResponseWriter, req *http.Request) {
	logger := api.Logger()
	session, err := api.Session()
	if err != nil {
		logger.Error("se", "unable to retrieve session", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	secretProvider, err := api.SecretProvider()
	if err != nil {
		logger.Error("se", "unable to retrieve secret provider", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	metadata := api.Metadata()
	idpName := metadata["idpName"].(string)
	ldapURL := metadata["ldapURL"].(string)
	ldapServerName := metadata["ldapServerName"].(string)
	ldapBaseDN := metadata["ldapBaseDN"].(string)
	ldapCASecretName := metadata["ldapCASecretName"].(string)

	username, password, ok := req.BasicAuth()
	if !ok {
		logger.Error("se", "unable to read basic auth headers")
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	logger.Info(
		"se", "logging user in",
		"req", req.URL.String(),
		"username", username,
	)
	certPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Error("se", "unable to load system cert pool", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	ldapCACert := secretProvider.GetString(ldapCASecretName)
	certPool.AppendCertsFromPEM([]byte(ldapCACert))
	tlsCfg := &tls.Config{
		ServerName: ldapServerName,
		RootCAs:    certPool,
	}

	conn, err := ldap3.DialURL(ldapURL, ldap3.DialWithTLSConfig(tlsCfg))
	if err != nil {
		logger.Error("se", "unable to do ldap dial", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	usernameKey := fmt.Sprintf("uid=%s,%s", username, ldapBaseDN)
	err = conn.Bind(usernameKey, password)
	if err != nil {
		logger.Error("se", "unable to do ldap bind", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_ = session.SetString(fmt.Sprintf("%s.authenticated", idpName), "true")
	_ = session.SetString(fmt.Sprintf("%s.cn", idpName), username)

	err = session.Save()
	if err != nil {
		logger.Error("se", "unable to save session", "error", err.Error())
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	logger.Info("se", "user successfully authenticated", "username", username)
	return
}
