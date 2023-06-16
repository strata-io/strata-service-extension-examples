# LDAP Search Service Extension

There are times when we need to construct a more unique LDAP search query that isn't
easily available via the LDAP connector. For example, if you need to make a query
that returns multiple result entries.

This example, queries LDAP securely by upgrading TCP the connection to TLS and making
a request to it to retrieve groups with a specific `uniqueMember` attribute. 
The structure of the LDAP can be seen in the [`example.ldif`](./example.ldif) file.

For more information regarding the `maverics/ldap` pkg, please refer to
the [Service Extension Maverics package documentation][maverics-ldap-docs].

## Setup

Please reference the [maverics.yaml](maverics.yaml) configuration file and the
Service Extension files for a set of TODOs. These action items are the minimum amount
of changes necessary to get the example running.

## Testing

1. Restart the Orchestrator after the TODOs are resolved and ensure it starts
   successfully.
1. Navigate to the URL the Orchestrator is listening on in you browser. If the
   Orchestrator was running on your laptop, the URL would
   be https://localhost/headers.
1. You should now be redirected to Azure and prompted for authentication.
1. After successfully logging in, the custom headers that were created will be sent
   to the upstream application. This can be confirmed by verifying the
   `EXAMPLE-GROUPS` header is rendered on the `/headers` page of the sample app.

[maverics-ldap-docs]: https://docs.strata.io/orchestrator-reference/service-extensions/maverics-packages#package-maverics-ldap