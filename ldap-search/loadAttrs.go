package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"

	ldap3 "github.com/go-ldap/ldap/v3"
	"github.com/strata-io/service-extension/orchestrator"
	"github.com/strata-io/service-extension/secret"
)

// LoadAttrs loads attributes from LDAP and then stores them on the session for later
// use.
func LoadAttrs(api orchestrator.Orchestrator, _ http.ResponseWriter, _ *http.Request) error {
	logger := api.Logger()
	session, err := api.Session()
	if err != nil {
		return fmt.Errorf("unable to retrieve session: %w", err)
	}
	secretProvider, err := api.SecretProvider()
	if err != nil {
		return fmt.Errorf("unable to get secret provider: %w", err)
	}

	metadata := api.Metadata()
	ldapServerName := metadata["ldapServerName"].(string)
	ldapBaseDN := metadata["ldapBaseDN"].(string)
	ldapFilterFmt := metadata["ldapFilterFmt"].(string)
	delimiter := metadata["delimiter"].(string)

	logger.Info("se", "loading attributes from LDAP")

	uid, err := session.GetString("azure.email")
	if err != nil {
		return fmt.Errorf("failed to find user email required for LDAP query: %w", err)
	}

	filter := fmt.Sprintf(ldapFilterFmt, uid)
	groupsMap, err := getGroups(ldapServerName, ldapBaseDN, filter, secretProvider)
	if err != nil {
		return fmt.Errorf("unable to get groups: %w", err)
	}

	groups := make([]string, 0, len(groupsMap))
	for k, _ := range groupsMap {
		groups = append(groups, k)
	}

	list := strings.Join(groups, delimiter)
	logger.Debug(
		"se", "setting groups attribute on session",
		"se.groups", list,
	)

	err = session.SetString("se.groups", list)
	if err != nil {
		return fmt.Errorf("unable to set 'se.groups' in session: %w", err)
	}
	err = session.Save()
	if err != nil {
		return fmt.Errorf("unable to save session: %w", err)
	}

	return nil
}

// getGroups will search the LDAP ldapBaseDN using the provided filter and return a
// list of unique groups.
func getGroups(serverName, baseDN, filter string, secretP secret.Provider) (map[string]struct{}, error) {
	ldapURL := fmt.Sprintf("ldap://%s", serverName)
	conn, err := ldap3.DialURL(ldapURL)
	if err != nil {
		return nil, fmt.Errorf("unable to dial ldap3: %w", err)
	}
	defer conn.Close()

	caCert := secretP.GetString("ldapCACert")
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
		ServerName: serverName,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to start tls: %w", err)
	}

	serviceAccountUsername := secretP.GetString("serviceAccountUsername")
	serviceAccountPassword := secretP.GetString("serviceAccountPassword")
	err = conn.Bind(serviceAccountUsername, serviceAccountPassword)
	if err != nil {
		return nil, fmt.Errorf("unable to bind ldap3: %w", err)
	}

	searchReq := ldap3.NewSearchRequest(
		baseDN, // The base dn to search
		ldap3.ScopeWholeSubtree, ldap3.NeverDerefAliases, 0, 0, false,
		filter,         // The filter to apply
		[]string{"cn"}, // A list attributes to retrieve
		nil,
	)
	searchResult, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("unable to search ldap3: %w", err)
	}

	groups := make(map[string]struct{})
	for _, entry := range searchResult.Entries {
		groups[entry.GetAttributeValue("cn")] = struct{}{}
	}

	return groups, nil
}
