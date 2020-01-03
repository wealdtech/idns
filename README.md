# idns

[![Tag](https://img.shields.io/github/tag/wealdtech/idns.svg)](https://github.com/wealdtech/idns/releases/)
[![License](https://img.shields.io/github/license/wealdtech/idns.svg)](LICENSE)
[![Travis CI](https://img.shields.io/travis/wealdtech/idns.svg)](https://travis-ci.org/wealdtech/idns)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/idns.svg)](https://codecov.io/github/wealdtech/idns)

IPFS/DNS integration helper, retrieving DNS zonefiles from IPFS based on ENS records.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`idns` is a standard Go binary which can be installed with:

```sh
go get github.com/wealdtech/idns
```

## Usage

`idns` takes three parameters that must be specified:

  - `connection` a connection to an Ethereum node, for example `/home/me/.ethereum/geth.ipc`
  - `dir` the directory in which the resultant zonefiles will be written, for example `/home/dns/zones`
  - `gateway` the IPFS gateway from which to obtain zonefiles, for example `https://ipfs.infura.io/`

In addition there is one optional parameter:

  - `from` the block from which to start querying for relevant changes to ENS.  This should usually be left blank

### Zone validation

When `idns` retrieves a zonefile from IPFS it carries out the following checks to ensure the zonefile is valid:

  - parses the zonefile to ensure it does not contain any syntax errors
  - confirms the zonefile contains a single SOA record
  - confirms the domain named in the SOA record is the same as the domain from which the DNS update event originated
  - confirms the address of the contract which sent the DNS update event is the same as the ENS resolver of the DNS zone

### Output

`ìdns` stores the zonefile in the directory given by the `dir` parameter.  The file name for the zone is the domain name prefixed with `db.`, so for example the output file for zone `example.com` would be `db.example.com`.

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/idns/issues).

## License

[Apache-2.0](LICENSE) © 2019 Weald Technology Trading Ltd
