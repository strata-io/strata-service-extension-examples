# LDAP Search Service Extension

There are times when you may find the need to construct a unique LDAP search
query that isn't easily available via the LDAP connector. For example,
returning multiple result entries.

This example will show you how to query LDAP securely by upgrading the TCP
connection to TLS and then make a request to the LDAP server to retrieve groups
with a specific `uniqueMember` attribute.

The structure of the LDAP can be seen in the [`example.ldif`](./example.ldif) file.

For more information regarding the `maverics/ldap` pkg, please refer to
the [Service Extension Maverics package documentation][maverics-ldap-docs].

## Setup

Please reference the [maverics.yaml](maverics.yaml) configuration file and the
Service Extension files it references for a set action items specified with`TODO`.
These action items are changes necessary to get this example running.

## Testing

1. Complete all the action items specified by `TODO`s
1. Restart the Orchestrator, and ensure it starts successfully.
1. Navigate to the URL the Orchestrator is listening on in your browser:
   e.g. https://localhost/headers.
1. You should now be redirected your specified IDP and prompted for authentication.
1. After successfully logging in, the custom headers that were created will be sent
   to the upstream application. This can be confirmed by verifying the
   `EXAMPLE-GROUPS` header is rendered on the `/headers` page of the sample app.

[maverics-ldap-docs]: https://docs.strata.io/orchestrator-reference/service-extensions/maverics-packages#package-maverics-ldap
