# Masked Fastmail

A simple CLI tool for managing [Fastmail masked email aliases](https://www.fastmail.com/features/masked-email/).
Easily create new aliases for websites or manage existing ones.

## Features

- Get or create masked email addresses for domains
- Aliases are automatically copied to clipboard
- Enable/disable/delete aliases

## Usage

![demo](./demo.gif)

```text
Usage:
  masked_fastmail <url>   (no flags)
  manage_fastmail <alias> [flags]

Flags:
      --delete    delete alias (bounce messages)
  -d, --disable   disable alias (send to trash)
  -e, --enable    enable alias
  -h, --help      show this message
  -v, --version   show version information
```

The following environment variables must be set:

```shell
export FASTMAIL_ACCOUNT_ID=your_account_id
export FASTMAIL_API_KEY=your_api_key
```

### Examples

#### Get or create alias

A new alias will only be created if one does not already exist.
In either case, the alias is automatically copied to the clipboard.[^1]

[^1]: Copying is done with [Clipboard for Go](https://pkg.go.dev/github.com/atotto/clipboard#section-readme) and should work on all platforms.

```shell
masked_fastmail example.com
```

#### Enable an existing alias

New Fastmail aliases are initialized to `pending`, and are set to `enabled` once they receive their first email.
However, they get automatically deleted if no email is received within 24 hours.
Some services may not send a timely welcome email, in which case it's helpful to manually enable the alias.

```shell
masked_fastmail --enable user.1234@fastmail.com
```

#### Disable an alias

This causes all new new emails to be moved to trash.

```shell
masked_fastmail --disable user.1234@fastmail.com
```

## Installation

### Option 1: Download a pre-built binary

Download the latest release from the [releases page](https://github.com/fredrikrab/masked_fastmail/releases).

### Option 2: Use `go install`

```shell
go install github.com/fredrikrab/masked_fastmail@latest
```

### Option 3: Build from source

1. Clone the repository
2. Run `go build -o masked_fastmail`

#### Prerequisites

- Go 1.22+
- Fastmail API credentials

#### API documentation

- The API documentation can be found at [https://www.fastmail.com/dev/maskedemail](https://www.fastmail.com/dev/maskedemail)
- It's also helpful to review the [JMAP protocol](https://jmap.io/crash-course.html)

## License

BSD 3-Clause License
