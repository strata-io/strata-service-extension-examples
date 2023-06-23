# LDAP TLS Service Extension

This example will show you how to dial and bind to an LDAP directly over TLS without
having to upgrade the TCP connection to TLS.

For more information regarding the `maverics/ldap` pkg, please refer to
the [Service Extension Maverics package documentation][maverics-ldap-docs].

## Setup

Please reference the [maverics.yaml](maverics.yaml) configuration file and the
Service Extension files it references for a set action items specified with`TODO`.
These action items are changes necessary to get this example running.

## Testing

1. Complete all the action items specified by `TODO`s.
1. Restart the Orchestrator, and ensure it starts successfully.
1. Use an HTTP client such as Postman or Curl to make a request to the orchestrator
   `https://localhost/` with the username and password of the LDAP user as Basic Auth
   headers.

[maverics-ldap-docs]: https://docs.strata.io/orchestrator-reference/service-extensions/maverics-packages#package-maverics-ldap
