# Amazon Verified Permissions Quick Start

Overview blah blah

## Setup

Please reference the [maverics.yaml](maverics.yaml) configuration file for a set of
TODOs. These action items are the minimum amount of changes necessary to get the
example running.

## Testing

1. Restart the Orchestrator after the TODOs are resolved and ensure it starts successfully.
1. Navigate to the URL the Orchestrator is listening on in you browser. If the
Orchestrator was running on your laptop, the URL would be https://localhost/headers.
1. You should be redirected for authentication against the configured IDP.
1. After successful auth, application traffic will be proxied through the AppGateway 
and the attributes returned from the API will be sent to app as HTTP headers. The
custom headers that are defined in the configuration file will be sent to the 
upstream application.