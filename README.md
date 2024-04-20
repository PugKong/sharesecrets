# ShareSecrets

[![build](https://github.com/PugKong/sharesecrets/actions/workflows/ci.yml/badge.svg)](https://github.com/PugKong/sharesecrets/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/PugKong/sharesecrets/graph/badge.svg?token=3UNSPY19XI)](https://codecov.io/github/PugKong/sharesecrets)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/PugKong/sharesecrets)](https://goreportcard.com/report/github.com/PugKong/sharesecrets)
[![Release](https://img.shields.io/github/release/PugKong/sharesecrets.svg?style=flat-square)](https://github.com/PugKong/sharesecrets/releases/latest)

A web-based service designed to share sensitive information (secrets) with others in a more secure way than traditional chat messages.
With ShareSecrets, you can share your secrets and control how long they are available before being destroyed.

The service consists of two pages.

One page allows you to share your secret with a passphrase and lifetime. After a successful secret creation you'll obtain the link to share.

The other page allows you to open a secret using the passphrase. Note that the secret will be destroyed

- After successful opening
- After three unsuccessful attempts to open

## Usage

Docker

```sh
$ docker run --rm -p 8000:8000 ghcr.io/pugkong/sharesecrets:master
```
