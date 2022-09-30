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
You'll need to [create a FastMail API token](https://beta.fastmail.com/settings/security/tokens).

> **🔒 The only necessary scope is "Masked Email".**
>
> Always use unique API tokens with the minimum scope(s) necessary for different purposes.

You can test authentication by running `maskedemail-cli -token abcdef12345 session`.

## Usage

```
Usage of maskedemail-cli:
Flags:
  -accountid string
    	fastmail account id (or MASKEDEMAIL_ACCOUNTID env)
  -appname string
    	the appname to identify the creator (or MASKEDEMAIL_APPNAME env) (default: maskedemail-cli)
  -token string
    	the token to authenticate with (or MASKEDEMAIL_TOKEN env)

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

$ maskedemail-cli -token abcdef12345 list | grep facebook
123@mydomain.com     https://www.facebook.com       disabled   2022-08-09T07:49:43Z
```

## Other resources and things powered by this CLI 

- [Siri Shortcut](https://www.icloud.com/shortcuts/973a2453b95d4dab97db950260283f4d) to disable the masked email of the currently selected message in Apple Mail
- [maskedemail-js](https://github.com/dvcrn/maskedemail-js): Node package ready to import, backed by this CLI compiled to wasm

## License

MIT

## Attributions

JMAP API documentation from [jmapio/jmap][] (Apache 2.0 / Copyright 2016 Fastmail Pty Ltd)

[jmapio/jmap]: https://github.com/jmapio/jmap
