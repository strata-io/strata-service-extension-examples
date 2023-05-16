# Amazon Verified Permissions Quick Start
# Amazon Verified Permissions Quick Start Guide

## Download assets from the Strata public repository
Strata has made a number of assets available for download on [our Github repository](https://github.com/strata-io/strata-service-extension-examples/amazon-verified-permissions) to streamline the Maverics setup process, including a readme with more detailed instructions. To continue this setup, you will need to download the following files to a directory on your machine:

* Sample Cedar policy: The sample Amazon Verified Permissions policy in Cedar.
* **Amazon_Recipe.json**: The custom configuration that will be copied to Maverics.
* **amazon-verified-permissions.go**: The code for the service extension that will connect your user flow to your Amazon Verified Permissions policy.
* **maverics.env**: The file for your local environment
* **localhost.cer** and **localhost.key** files: The certificates to run your local Orchestrator


## Create the Amazon Verified Permissions policy and an IAM user
To implement Maverics with Amazon Verified Permissions, you must first create your Amazon Verified Permissions policies in the AWS Management Console. Maverics will use these policies to perform the authorization.

<Link to the AVP docs once it’s available>

Use the sample policy available in our Github repository. This  sample policy permits view and create access to an application for a user called `aadkins@sonar-systems.com`.

```
permit (
	principal == User::"aadkins@sonarsystems.com",
	action in [Action::"create", Action::"view"],
	resource == Endpoint::"/"
);
```

Now, click Settings in the sidebar of Amazon Verified Permissions. Make a note of the Policy Store ID, as this will be used in the service extension.

Additionally, you will need to [create a separate IAM user](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html). For this user, create a new policy with the code block below. Name this policy Sonar and select this policy for the user:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "verifiedpermissions:*"
            ],
            "Resource": "*"
        }
    ]
}
```

Create an access key for this user, and make a note of the AWS key ID and secret key, as this will also be used in the service extension.

## Install the Sonar demo app
Sonar is an application provided by Strata to demonstrate the user flow. The app is stored in a Docker container and can be run with the following steps:

1. Go to [Docker.com](https://docker.com) and download the version of Docker Desktop for your operating system. 
1. Follow the steps to install Docker Desktop.
1. After installation, open a command prompt and run the following command to download and run the Sonar demo app:
`docker run -p 8987:8987 strataidentity/sonar sonar`

## Sign up for Maverics 
Next, sign up for the Maverics Identity Orchestration Platform at http://maverics.strata.io. You can sign up with HYPR, Google, or Microsoft Azure SSO. 

After signing up for an account, you will land on the Dashboard. You can then set up your orchestrator, environment, identity fabric, and user flow using the buttons on the screen or left navigation.

## Create a local environment for testing
Environments are storage locations that contain user flow configurations for your applications. You can create multiple environments (for example: dev, test, staging, and production environments), configure storage containers, and assign an orchestrator to each environment.

1. From the sidebar, go to **Environments** and click the + icon next to  Download Only/Local.
1. Configure the following:
	* **Name**: A friendly name for your environment. For this example, let’s use local-environment.
	* **Description**: Additional description of the environment.
1. Click **Create**.
1. Go to **Environments** from the left navigation, and click the environment you’ve just created.
1. Download the appropriate Orchestrator file for your operating system, as well as the public key. Save these files to the same directory as the certificates and environment file from our Github repo.
1. Follow the instructions to install based on your operating system:
	* ​[Windows](https://docs.strata.io/maverics-orchestrator/install-and-setup/install-windows)​
	* ​[Linux](https://docs.strata.io/maverics-orchestrator/install-and-setup/install-linux)​
	* ​[Docker](https://docs.strata.io/maverics-orchestrator/install-and-setup/install-docker)​
	* ​[Mac](https://docs.strata.io/maverics-orchestrator/install-and-setup/install-macos)​
1. Start the Orchestrator based on the instructions in the installation.
	* Optionally, you can use remote configuration to connect it to a shared storage provider. For the purposes of this evaluation, we will use local storage only.

The Orchestrator instance will then attempt to read the configuration from your local storage, but it will fail until you've deployed the Orchestrator in the next section.

## Import the demo recipe
Strata provides a Maverics recipe named Amazon-Verified-Permissions-Recipe.json for evaluation purposes. This recipe automatically configures Amazon Cognito as an identity provider, Sonar as an application, and a user flow that connects the two.

To upload this recipe:

1. Go to the dashboard and click the **Upload custom configuration** button at the top of the screen. 
1. From the **Import** screen, enter a name for the user flow.
1. Copy the code from Amazon_Recipe.json and paste it into the Configuration text box.
1. Click **Create**.

## Set up Amazon Verified Permissions as a service extension
Now you can set up Amazon Verified Permissions as a service extension. 

1. First, open the **amazon-verified-permissions.go** file downloaded from the Github repository with a code editor, and copy the raw code. 
1. Go to the **Service Extensions** page from the left navigation in Maverics, and select **Authorization Service Extension**.
1. Provide the name Amazon Verified Permissions and a description, and click **Create**.
1. When you click Create, the service extension code box appears. Paste the code copied from the **amazon-verified-permissions.go** file. Follow the instructions in the code to replace the Policy Store ID and keys with the values you collected from the Amazon Admin Console.
1. Click **Update** to save the code. 
1. At the bottom of the screen under Providers, select your Cognito identity provider instance and click **Add**. Then click **Update** again.

## Connect your user flow to Amazon Verified Permissions
From here, you can now complete the setup of your user flow. 

1. Go to the the **User Flows** page from the left navigation. The Sonar user flow from the demo recipe is listed.
1. Click the user flow. 
1. On the next screen, you should see **Sonar** under Application, and **Amazon Cognito** already selected as the authentication provider.

	1. 	Under **Add access control policy**, select a resource location (configured from the recipe), and click **Add**. 
	1. 	You will then be prompted to configure access control. Under Authentication, select **Require authentication by Amazon_Cognito**. Under Access Controls, select **Use service extension: Amazon Verified Permissions policy**. Click **Update**.
	1. Defining additional headers is optional. To do this, write the attribute under Headers, select the attribute provider, define the claim, and click **Add**.
	1. When you are done configuring your user flow, click **Update**.
1. In the Latest Revision section, the indicator should notify you that your user flow has been updated. Click **Save Revision.**
1. Finally, click **Publish** in the upper right corner to deploy your user flow to your environment and start using Maverics.

## Test your user flow
You can test your user flow by logging into the Sonar app.

1. Open a browser window to access the Sonar app at [https://localhost](https://localhost)
1. Login as `aadkins@sonar-systems.com` with the password `password`.
1. This user should be denied.
1. Go to the Amazon Verified Permissions policy page and change the policy from forbid to permit and save the policy.
1. Open a new browser window and repeat steps 1 and 2.

To see this in action and for instructions on how to test your user flows, watch our demonstration video of the complete app modernization process. For more detailed information on setting up Maverics, view our documentation at [docs.strata.io](https://docs.strata.io).