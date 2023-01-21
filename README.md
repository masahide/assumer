# assumer
AWS assume role credential wrapper.

Implemented by golang with reference to [aswrap](https://github.com/fujiwara/aswrap).

[![Go Report Card](https://goreportcard.com/badge/github.com/masahide/assumer)](https://goreportcard.com/report/github.com/masahide/assumer)

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



