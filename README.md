# Fastmail MaskedEmail CLI

CLI to create Fastmail Masked Emails for whenever you need to

![showcase](./showcase.gif)

## Setup

```
go get github.com/dvcrn/maskedemail-cli
```

or newer Go versions

```
go install github.com/dvcrn/maskedemail-cli@latest
```

### Authentication
You'll need to [create a FastMail API token](https://www.fastmail.com/settings/security/tokens?u=1eb14002).

> **🔒 The only necessary scope is "Masked Email".**
>
> Always use unique API tokens with the minimum scope(s) necessary for different purposes.

You can test authentication by running `maskedemail-cli session`.

## Usage

Currently only the `create` command is implemented

```
Usage of maskedemail-cli
Flags:
  -accountid string
        fastmail account id
  -appname string
        the appname to identify the creator (default "maskedemail-cli")
  -token string
        the token to authenticate with

Commands:
  maskedemail-cli create <domain>
  maskedemail-cli enable <masked email>
  maskedemail-cli disable <masked email>
  maskedemail-cli session
  maskedemail-cli list

```

Example:

```
$ maskedemail-cli -token abcdef12345 create facebook.com
$ maskedemail-cli -token abcdef12345 enable 123@mydomain.com
$ maskedemail-cli -token abcdef12345 disable 123@mydomain.com
```

## License

MIT

## Attributions

JMAP API documentation from [jmapio/jmap][] (Apache 2.0 / Copyright 2016 Fastmail Pty Ltd)

[jmapio/jmap]: https://github.com/jmapio/jmap
