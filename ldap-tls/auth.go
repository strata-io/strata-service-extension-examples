package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"

	"maverics/app"
	"maverics/ldap"
	"maverics/log"
	"maverics/secret"
	"maverics/session"
)

const (
	idpName          = "ldap"
	ldapURL          = "ldap://ldap.examples.com"
	ldapServerName   = "ldap.examples.com"
	ldapBaseDN       = "ou=People,dc=examples,dc=com"
	ldapCASecretName = "ldapCA"
)

// IsAuthenticated determines if the user has been authenticated.
func IsAuthenticated(ag *app.AppGateway, rw http.ResponseWriter, req *http.Request) bool {
	log.Debug("se", "determining if user is authenticated")

	if session.GetString(req, fmt.Sprintf("%s.authenticated", idpName)) == "true" {
		return true
	}

	return false
}

// Authenticate authenticates the user against the IDP.
func Authenticate(ag *app.AppGateway, rw http.ResponseWriter, req *http.Request) error {
	username, password, ok := req.BasicAuth()
	if !ok {
		log.Error("se", "unable to read basic auth headers")
		return errors.New("unable to read basic auth headers")
	}

	log.Info(
		"se", "logging user in",
		"req", req.URL.String(),
		"username", username,
	)
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("unable to load system cert pool: %w", err)
	}
	ldapCACert := secret.GetString(ldapCASecretName)
	certPool.AppendCertsFromPEM([]byte(ldapCACert))
	tlsCfg := &tls.Config{
		ServerName: ldapServerName,
		RootCAs:    certPool,
	}

	conn, err := ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(tlsCfg))
	if err != nil {
		return fmt.Errorf("unable to do ldap dial: %w", err)
	}
	defer conn.Close()

	usernameKey := fmt.Sprintf("uid=%s,%s", username, ldapBaseDN)
	err = conn.Bind(usernameKey, password)
	if err != nil {
		return fmt.Errorf("unable to do ldap bind: %w", err)
	}
	session.Set(req, fmt.Sprintf("%s.authenticated", idpName), "true")
	session.Set(req, fmt.Sprintf("%s.cn", idpName), username)
	log.Info("se", "user authenticated", "username", username)
	return nil

	log.Info("se", "successfully authenticated", "username", username)
	return nil
}
