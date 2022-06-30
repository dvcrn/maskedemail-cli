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

You have 2 ways to authenticate with the fastmail API

#### 1. Username + Password

The easiest is to use the built-in auth command to get your access token + accountid. However the token does eventually expire and you'll need to re-auth, so I'd recommend using the 1Password method explained in 2. instead

```
$ maskedemail-cli auth <email> <password>
authentication successful!
accountID:  xxxx
token:  yyyy
```

#### 2. Extracting the refresh token from 1Password

This is technically the better method because the refresh-token does not expire.
Use a reverse proxy like Proxyman, mitmproxy or charles, then start your browser and create a masked email

Find the refresh token inside the body when the 1Password extension first connects to the fastmail API and write that down

Specify the token as usual with the `-token` flag, but also set `-refresh` to tell the CLI that the token is a refresh token.

#### Why is auth so complicated??

The Masked Email capability is not available through the normal JMAP API yet, so if we were to create a token with the JMAP API scope, it wouldn't be able to use Masked Emails.

I contacted the fastmail team and while there are plans to move this into general availablity, it's not gonna be anytime soon (though likely this year).

## Usage

Currently only the `create` command is implemented

```
Usage of maskedemail-cli
Flags:
  -accountid string
        fastmail account id
  -appname string
        the appname to identify the creator (default "maskedemail-cli")
  -refresh
        whether the token is a refresh token
  -token string
        the token to authenticate with

Commands:
  maskedemail-cli create <domain>
  maskedemail-cli auth <email> <password>

```

Example:

```
$ maskedemail-cli -accountid xxxx -token abcdef12345 create facebook.com
```

## License

MIT
