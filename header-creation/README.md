# Header Creation Service Extension

For header-based applications, there is a common need to construct custom HTTP 
headers. For example, custom headers may be a combination of identity attributes into
a single value, or a single attribute that has been parsed into a certain format.

## Setup
Please reference the [maverics.yaml](maverics.yaml) configuration file for a set of
TODOs. These action items are the minimum amount of changes necessary to get the
example running.

## Testing
1. Restart the Orchestrator after the TODOs are resolved and ensure it starts successfully.
1. Navigate to the URL the Orchestrator is listening on in you browser. If the 
Orchestrator was running on your laptop, the URL would be https://localhost/headers.
1. You should now be redirected to Azure and prompted for authentication. 
1. After successfully logging in, the custom headers that were created will be sent
to the upstream application. This can be confirmed by verifying the custom 
`EXAMPLE-FIRST-NAME` and `EXAMPLE-LAST-NAME` headers are rendered on the `/headers` 
page of the sample app.