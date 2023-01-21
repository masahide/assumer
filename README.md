# assumer

A simple program to make it easy to use AWS Single Sign On (AWS SSO) credentials and credentials to assume roles in ~/.aws/credentials with tools that do not recognize AWS profile entries.

It provides the following functionality
- Execute commands to assume roles in ~/.aws/credentials
- Assume roles and execute commands via AWS SSO

## Assume role in `~/.aws/credentials` and execute command

The `~/.aws/credentials` file should look like this:
```toml
[my-profile]
aws_access_key_id=XXX
aws_secret_access_key=YYY

[foo]
region=ap-northeast-1
source_profile=my-profile
role_arn=arn:aws:iam::999999999999:role/MyRole
```

Then you can run:
```bash
AWS_PROFILE=foo assumer <some_command> [arg1 arg2...]
```

## Assume roles and execute commands via AWS SSO

The ~/.aws/config file should look like this:
```toml
[default]
sso_start_url = xxxxxxxxxxxx
sso_region = us-west-2
sso_account_id = xxxxxxxxxxxx
sso_role_name = SSORoleName

[profile account1]
role_arn = arn:aws:iam::xxxxxxxxxxxx:role/role-to-be-assumed
source_profile = default
region = ap-northeast-1
```

Then you can run:
```bash
AWS_PROIFILE=account1 assumer <some_command> [arg1 arg2...]
```
Execute `<command>` with the `role-to-be-assumed` role.
