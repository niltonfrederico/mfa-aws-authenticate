# MFA AWS Authenticate

> MFA AWS Authenticate is a utility to facilitate using MFA in a terminal using `aws-cli`

## 💻 Prerequisites

* You got to have [go](https://go.dev/doc/install) installed
* You got to have [aws-cli](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) (v1 or v2 should be fine).
* You need to have `iam:ListMFADevices` permission on aws.

## 🚀 Installing

* Download this repository and run

```bash
go build -o build/mfa-aws github.com/niltonfrederico/mfa-aws-authenticate
mv build/mfa-aws /usr/bin/mfa-aws
```

## ☕ Using MFA AWS Authenticate

To use `mfa-aws` simply run:

```bash
mfa-aws auth --code {your_mfa_code}
```

## 📫 Contributing

For now just forkit and open your pull request!

## 📝 License

This project is under license. See the file [LICENSE](LICENSE.md) for more details.
