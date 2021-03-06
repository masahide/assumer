# assumer
AWS assume role credential wrapper.

Implemented by golang with reference to [aswrap](https://github.com/fujiwara/aswrap).

[![Go Report Card](https://goreportcard.com/badge/github.com/masahide/assumer)](https://goreportcard.com/report/github.com/masahide/assumer)
[![Build Status](https://travis-ci.org/masahide/assumer.svg?branch=master)](https://travis-ci.org/masahide/assumer)
[![codecov](https://codecov.io/gh/masahide/assumer/branch/master/graph/badge.svg)](https://codecov.io/gh/masahide/assumer)
[![goreleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square)](https://github.com/goreleaser)

## Description

assumer is useful for some commands which couldn't resolve an assume role credentials in ~/.aws/credentials and ~/.aws/config.

For example,

- Implemented with [aws-sdk-go](https://github.com/aws/aws-sdk-go)
- [Terraform](https://www.terraform.io/)
- [Packer](https://www.packer.io/)
- [Ansible](https://www.ansible.com)
- etc.

## Installation

see [releases page](https://github.com/masahide/assumer/releases).


## Usage

```bash
$ AWS_PROFILE=foo assumer some_command [arg1 arg2...]
```

Refer to "[aswrap](https://github.com/fujiwara/aswrap/blob/master/README.md#usage)" because it is the same "[usage as aswrap](https://github.com/fujiwara/aswrap/blob/master/README.md#usage)"


## Clearing Cached Credentials

When you assume a role, the `assumer` caches the temporary credentials locally until they expire. If your role's temporary credentials are revoked, you can delete the cache to force the `assumer` to retrieve new credentials.


```bash
rm -rf ~/.assumer/cache
```
