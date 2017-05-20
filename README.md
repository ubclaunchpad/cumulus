# Cumulus

Crypto-currency that doesn't waste your time.  

[![ZenHub](https://raw.githubusercontent.com/ZenHubIO/support/master/zenhub-badge.png)](https://zenhub.com)
[![Coverage Status](https://coveralls.io/repos/github/ubclaunchpad/cumulus/badge.svg?branch=dev)](https://coveralls.io/github/ubclaunchpad/cumulus?branch=dev)
[![Build Status](https://travis-ci.org/ubclaunchpad/cumulus.svg?branch=dev)](https://travis-ci.org/ubclaunchpad/cumulus)

## Installation

Install dependencies. We need to manually use version 0.11.1 of Glide temporarily because 0.12 introduced a bug in recursive dependencies.

Install Glide.
```sh
make install-glide
```

Get dependencies.
```sh
make deps
```

Build.
```sh
make
```

## Testing

```
make test
```
