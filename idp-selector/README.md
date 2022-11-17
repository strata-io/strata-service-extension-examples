# IDP Selector Service Extension

Certain enterprises have the need to support multiple IDPs simultaneously in order
to facilitate authentication and authorization. This need may occur as the result of
an acquisition, or as part of a migration project to a new IDP.

This example uses Azure Active Directory and Auth0 as the IDPs. However, modifying
the example to use any set of IDPs should be straightforward.

## Setup
Please reference the [maverics.yaml](maverics.yaml) configuration file for a set of
TODOs. These action items are the minimum amount of changes necessary to get the 
example running. 

Additionally, the [`auth.go`](auth.go) Service Extension contains documentation on 
the implementation. No changes are required in that file, however.

## Testing
1. Restart the Orchestrator after the TODOs are resolved and ensure it starts successfully.
1. Navigate to the URL the Orchestrator is listening on in you browser. If the 
Orchestrator was running on your laptop, the URL would be https://localhost. 
1. You should now see a form asking which IDP to authenticate against. After 
selecting an IDP, you will be redirected for authentication. 
1. After successful auth, application traffic will be proxied through the AppGateway.

