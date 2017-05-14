# Cumulus

Crypto-currency that doesn't waste your time.  

[![ZenHub](https://raw.githubusercontent.com/ZenHubIO/support/master/zenhub-badge.png)](https://zenhub.com)
[![Coverage Status](https://coveralls.io/repos/github/ubclaunchpad/cumulus/badge.svg?branch=dev)](https://coveralls.io/github/ubclaunchpad/cumulus?branch=dev)
[![Build Status](https://travis-ci.org/ubclaunchpad/cumulus.svg?branch=dev)](https://travis-ci.org/ubclaunchpad/cumulus)

## Installation

Install dependencies. We need to manually use version 0.11.1 of Glide temporarily because 0.12 introduced a bug in recursive dependencies.

Install Glide.
```sh
go get github.com/Masterminds/glide
cd $GOPATH/src/github.com/Masterminds/glide
git checkout tags/v0.11.1
go install
```

Verify you have the correct version installed.
```sh
glide --version
```

Get dependencies.
```sh
cd $GOPATH/src/github.com/ubclaunchpad/cumulus
glide install
```

## Testing

```
go test ./...
```
