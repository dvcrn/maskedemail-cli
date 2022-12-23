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
  maskedemail-cli create [-domain "<domain>"] [-desc "<description>"] [-enabled=true|false (default true)]
  maskedemail-cli list [-show-deleted (default false)]
  maskedemail-cli enable <maskedemail>
  maskedemail-cli disable <maskedemail>
  maskedemail-cli delete <maskedemail>
  maskedemail-cli update -email <maskedemail> [-domain "<domain>"] [-desc "<description>"]
  maskedemail-cli session
```

Example:

```
$ maskedemail-cli -token abcdef12345 create  -domain "facebook.com" -desc "Facebook"
$ maskedemail-cli -token abcdef12345 enable 123@mydomain.com
$ maskedemail-cli -token abcdef12345 disable 123@mydomain.com

$ maskedemail-cli -token abcdef12345 list
Masked Email        For Domain     Description   State       Last Email At           Created At
123@mydomain.com    facebook.com   Facebook      disabled    2022-08-09T07:49:43Z    2022-07-02T06:34:21Z
```

## Other resources and things powered by this CLI 

- [Siri Shortcut](https://www.icloud.com/shortcuts/973a2453b95d4dab97db950260283f4d) to disable the masked email of the currently selected message in Apple Mail on macOS
- [maskedemail-js](https://github.com/dvcrn/maskedemail-js): Node package ready to import, backed by this CLI compiled to wasm
- [Masked Email Manager iOS App](https://apps.apple.com/us/app/masked-email-manager/id6443853807): iOS App backed by this CLI compiled to GopherJS

## License

MIT

## Attributions

JMAP API documentation from [jmapio/jmap][] (Apache 2.0 / Copyright 2016 Fastmail Pty Ltd)

[jmapio/jmap]: https://github.com/jmapio/jmap
