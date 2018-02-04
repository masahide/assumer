# assumer
AWS assume role credential wrapper.

[![Go Report Card](https://goreportcard.com/badge/github.com/masahide/assumer)](https://goreportcard.com/report/github.com/masahide/assumer)
[![Build Status](https://travis-ci.org/masahide/assumer.svg?branch=master)](https://travis-ci.org/masahide/assumer)
[![codecov](https://codecov.io/gh/masahide/assumer/branch/master/graph/badge.svg)](https://codecov.io/gh/masahide/assumer)

## Description

Implemented it with golang with reference to [aswrap](https://github.com/fujiwara/aswrap).

assumer is useful for some commands which couldn't resolve an assume role credentials in ~/.aws/credentials and ~/.aws/config.

For example,

- Implemented with [aws-sdk-go](https://github.com/aws/aws-sdk-go)
- [Terraform](https://www.terraform.io/)
- [Packer](https://www.packer.io/)
- [Ansible](https://www.ansible.com)
- etc.

## Installation

### Linux

For RHEL/CentOS:

```bash
sudo yum install https://github.com/masahide/assumer/releases/download/v0.1.4/assumer_amd64.rpm
```

For Ubuntu/Debian:

```bash
wget -qO /tmp/assumer_amd64.deb https://github.com/masahide/assumer/releases/download/v0.1.4/assumer_amd64.deb && sudo dpkg -i /tmp/assumer_amd64.deb
```

### macOS


install via [brew](https://brew.sh):

```bash
brew tap masahide/assumer https://github.com/masahide/assumer
brew install assumer
```


## Usage

Refer to "[aswrap](https://github.com/fujiwara/aswrap/blob/master/README.md#usage)" because it is the same "[usage as aswrap](https://github.com/fujiwara/aswrap/blob/master/README.md#usage)"


## Clearing Cached Credentials

When you assume a role, the `assumer` caches the temporary credentials locally until they expire. If your role's temporary credentials are revoked, you can delete the cache to force the `assumer` to retrieve new credentials.


```bash
rm -rf ~/.assumer/cache
```
