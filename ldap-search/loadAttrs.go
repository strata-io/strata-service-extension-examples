package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"maverics/app"
	"maverics/ldap"
	"maverics/log"
	"maverics/secret"
	"maverics/session"
)

const (
	// TODO: Adjust these values based on your LDAP configuration.
	ldapServerName = "ldap.examples.com"
	ldapBaseDN     = "dc=examples,dc=com"
	ldapFilterFmt  = "(&(uniquemember=uid=%s,ou=People,dc=examples,dc=com))"

	delimiter = ","
)

// LoadAttrs loads attributes from LDAP and then stores them on the session for later
// use.
func LoadAttrs(_ *app.AppGateway, _ http.ResponseWriter, req *http.Request) error {
	log.Debug("se", "loading attributes from LDAP")

	uid := session.GetString(req, "azure.email")
	if uid == "" {
		return fmt.Errorf("unable to get uid from session")
	}
	filter := fmt.Sprintf(ldapFilterFmt, uid)
	groupsMap, err := getGroups(ldapBaseDN, filter)
	if err != nil {
		log.Error("se", "unable to get groups", "error", err.Error())
		return err
	}

	groups := make([]string, 0, len(groupsMap))
	for k, _ := range groupsMap {
		groups = append(groups, k)
	}

	list := strings.Join(groups, delimiter)
	log.Debug(
		"se", "setting groups attribute on session",
		"se.groups", list,
	)

	session.Set(req, "se.groups", list)

	return nil
}

// getGroups will search the LDAP ldapBaseDN using the provided filter and return a
// list of unique groups.
func getGroups(baseDN, filter string) (map[string]struct{}, error) {
	ldapURL := fmt.Sprintf("ldap://%s", ldapServerName)
	log.Info("se", "dialing to ldap over tcp", "url", ldapURL)
	conn, err := ldap.DialURL(ldapURL)
	if err != nil {
		return nil, fmt.Errorf("unable to dial ldap: %w", err)
	}
	defer conn.Close()

	caCert := secret.GetString("ldapCACert")
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("unable to get system cert pool: %w", err)
	}
	ok := certPool.AppendCertsFromPEM([]byte(caCert))
	if !ok {
		return nil, errors.New("unable to append ca cert to pool")
	}
	err = conn.StartTLS(&tls.Config{
		RootCAs:    certPool,
		ServerName: ldapServerName,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to start tls: %w", err)
	}

	serviceAccountUsername := secret.GetString("serviceAccountUsername")
	serviceAccountPassword := secret.GetString("serviceAccountPassword")
	err = conn.Bind(serviceAccountUsername, serviceAccountPassword)
	if err != nil {
		return nil, fmt.Errorf("unable to bind ldap: %w", err)
	}

	searchReq := ldap.NewSearchRequest(
		baseDN, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,         // The filter to apply
		[]string{"cn"}, // A list attributes to retrieve
		nil,
	)
	searchResult, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("unable to search ldap: %w", err)
	}

	groups := make(map[string]struct{})
	for _, entry := range searchResult.Entries {
		groups[entry.GetAttributeValue("cn")] = struct{}{}

		fmt.Printf("%s: %v\n", entry.DN, entry.GetAttributeValue("cn"))
	}

	return groups, nil
}
