package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
)

const CACHE_PROFILE_NAME = "__cache-default"
const DEFAULT_PROFILE_NAME = "default"

func GetProfileToUse() string {
	fname := config.DefaultSharedCredentialsFilename()

	f, err := os.Open(fname)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	profile := ""
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if strings.Contains(line, CACHE_PROFILE_NAME) {
				profile = CACHE_PROFILE_NAME
				break
			} else if strings.Contains(line, DEFAULT_PROFILE_NAME) && profile == "" {
				profile = DEFAULT_PROFILE_NAME
			} else {
				panic("You need to have a [default] profile in your credentials file.")
			}
		}
	}
	f.Close()

	return profile
}

func ExecuteAwsConfigure(profile string, key string, value string) {
	exec.Command("aws", "configure", "--profile", profile, "set", key, value).Output()
}

func GetAwsConfig(profile string) aws.Config {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
	if err != nil {
		panic(err)
	}

	return awsConfig
}

func GetSessionToken(awsConfig aws.Config, code string) *sts.GetSessionTokenOutput {
	iamClient := iam.NewFromConfig(awsConfig)

	output, err := iamClient.ListMFADevices(context.TODO(), &iam.ListMFADevicesInput{})
	if err != nil {
		panic(err)
	}

	iamSerialNumber := output.MFADevices[0].SerialNumber
	stsClient := sts.NewFromConfig(awsConfig)

	sessionToken, err := stsClient.GetSessionToken(context.TODO(), &sts.GetSessionTokenInput{
		SerialNumber: iamSerialNumber,
		TokenCode:    &code,
	})
	if err != nil {
		panic(err)
	}

	return sessionToken
}

func GetCredentialsFromAwsConfig(awsConfig aws.Config) aws.Credentials {
	creds, err := awsConfig.Credentials.Retrieve(context.TODO())
	if err != nil {
		panic(err)
	}
	return creds
}

func main() {
	var Code string

	defer func() {
		rec := recover()
		if rec != nil {
			log.Fatal(rec)
		}
	}()

	var cmdAuthenticate = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate using a MFA code.",
		Run: func(cmd *cobra.Command, args []string) {
			var profile string = GetProfileToUse()
			var awsConfig aws.Config = GetAwsConfig(profile)
			var sessionToken *sts.GetSessionTokenOutput = GetSessionToken(awsConfig, Code)
			var credentials aws.Credentials = GetCredentialsFromAwsConfig(awsConfig)

			if profile != CACHE_PROFILE_NAME {
				ExecuteAwsConfigure(CACHE_PROFILE_NAME, "aws_access_key_id", credentials.AccessKeyID)
				ExecuteAwsConfigure(CACHE_PROFILE_NAME, "aws_secret_access_key", credentials.SecretAccessKey)
				ExecuteAwsConfigure(CACHE_PROFILE_NAME, "region", "us-east-1")
			}

			ExecuteAwsConfigure(DEFAULT_PROFILE_NAME, "aws_access_key_id", *sessionToken.Credentials.AccessKeyId)
			ExecuteAwsConfigure(DEFAULT_PROFILE_NAME, "aws_secret_access_key", *sessionToken.Credentials.SecretAccessKey)
			ExecuteAwsConfigure(DEFAULT_PROFILE_NAME, "aws_session_token", *sessionToken.Credentials.SessionToken)
		},
	}
	cmdAuthenticate.Flags().StringVarP(&Code, "code", "c", "", "MFA Code (required)")
	cmdAuthenticate.MarkFlagRequired("code")

	var rootCmd = &cobra.Command{
		Use:   "mfa-aws",
		Short: "mfa-aws is a wrapper to facilitate mfa authentication for aws cli v2.",
	}
	rootCmd.AddCommand(cmdAuthenticate)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
