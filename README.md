<p align="center">
  <img src="/.static/create-transaction.png" width="50%"/>
</p>

<div align="center">
  <h1>Cumulus</h1>
</div>

<p align="center">
  Crypto-currency that doesn't waste your time.  
</p>

<p align="center">
  <a href="https://travis-ci.org/ubclaunchpad/cumulus">
    <img src="https://travis-ci.org/ubclaunchpad/cumulus.svg?branch=dev" alt="Build Status" />
  </a>

  <a href="https://coveralls.io/github/ubclaunchpad/cumulus?branch=dev">
    <img src="https://coveralls.io/repos/github/ubclaunchpad/cumulus/badge.svg?branch=dev" alt="Coverage" />
  </a>

  <a href="https://godoc.org/github.com/ubclaunchpad/cumulus">
    <img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDocs" />
  </a>

  <a href="https://goreportcard.com/report/github.com/ubclaunchpad/cumulus">
    <img src="https://goreportcard.com/badge/github.com/ubclaunchpad/cumulus" alt="Clean code" />
  </a>

  <a href="https://zenhub.com">
    <img src="https://img.shields.io/badge/Shipping_faster_with-ZenHub-5e60ba.svg?style=flat" alt="Shipping faster with ZenHub" />
  </a>
</p>

<br>

## Introduction

At Launch Pad we’ve been interested in crypto-currency and blockchain tech for a while now, and after several months of experimentation we’re excited to announce our latest project. Cumulus is a new cryptocurrency with its own blockchain and token. The current command line interface allows users to create wallets, mine coins, and send funds to other users.

There are a lot of cryptocurrencies out there already, so it’s a fair question to ask what makes Cumulus special over other, more entrenched currencies like Bitcoin and Ethereum. The short answer is because we’re all excited by the possibilities created by this technology! In addition, there are many problems in the blockchain space that remain to be solved. Other cryptocurrencies have significant downsides like small block sizes and vast computation waste spent securing the network. We are addressing many of these problems in Cumulus. Beyond that, we see massive opportunities in this space in the years to come. Cumulus is the infrastructure on which we can build in the blockchain space as it matures.

## Building

First, [install Go](https://golang.org/doc/install#install) and grab the Cumulus source code:

```bash
$> go get -u github.com/ubclaunchpad/cumulus
```

Cumulus uses [dep](https://github.com/golang/dep) for dependency management. The following will install dep and run `dep ensure`:

```sh
make deps
```

You can now build and run Cumulus:

```sh
make run-console
```

## Testing

```
make test
```
