package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
)

func main() {
	var Code string

	var cmdAuthenticate = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate using a MFA code.",
		Run: func(cmd *cobra.Command, args []string) {
			fname := config.DefaultSharedCredentialsFilename()

			f, err := os.Open(fname)
			if err != nil {
				panic(err)
			}

			scanner := bufio.NewScanner(f)
			profile := ""
			hasCacheProfile := false
			for scanner.Scan() {
				line := scanner.Text()

				if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
					if strings.Contains(line, "__cache-default") {
						log.Print("Has cache default")
						profile = "__cache-default"
						hasCacheProfile = true
					} else if strings.Contains(line, "default") && profile == "" {
						profile = "default"
					}
				}
			}
			f.Close()

			cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			iamClient := iam.NewFromConfig(cfg)

			output, err := iamClient.ListMFADevices(context.TODO(), &iam.ListMFADevicesInput{})
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			iamSerialNumber := output.MFADevices[0].SerialNumber
			stsClient := sts.NewFromConfig(cfg)

			sessionToken, err := stsClient.GetSessionToken(context.TODO(), &sts.GetSessionTokenInput{
				SerialNumber: iamSerialNumber,
				TokenCode:    &Code,
			})
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			// cfg.Credentials.creds.v.AccessKeyID
			// cfg.Credentials.creds.v.SecretAccessKey
			creds, err := cfg.Credentials.Retrieve(context.TODO())
			if err != nil {
				log.Fatal(err)
			}

			if !hasCacheProfile {
				log.Print("Created cached profile")
				exec.Command("aws", "configure", "--profile", "__cache-default", "set", "aws_access_key_id", creds.AccessKeyID).Output()
				exec.Command("aws", "configure", "--profile", "__cache-default", "set", "aws_secret_access_key", creds.SecretAccessKey).Output()
				exec.Command("aws", "configure", "--profile", "__cache-default", "set", "region", "us-east-1").Output()
			}

			exec.Command("aws", "configure", "--profile", "default", "set", "aws_access_key_id", *sessionToken.Credentials.AccessKeyId).Output()
			exec.Command("aws", "configure", "--profile", "default", "set", "aws_secret_access_key", *sessionToken.Credentials.SecretAccessKey).Output()
			exec.Command("aws", "configure", "--profile", "default", "set", "aws_session_token", *sessionToken.Credentials.SessionToken).Output()
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
		fmt.Println(err)
		os.Exit(1)
	}
}
