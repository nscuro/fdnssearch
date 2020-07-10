# fdnssearch

![Build Status](https://github.com/nscuro/fdnssearch/workflows/Continuous%20Integration/badge.svg?branch=master)

## Installation

`GO111MOD=on go get -v github.com/nscuro/fdnssearch/cmd/fdnssearch`

fdnssearch requires Go >= 1.14

## Usage

```
$ fdnssearch -h
Usage:
  fdnssearch [flags]

Flags:
      --always-any            always search ANY dataset (ignored when -f is set)
  -c, --concurrency int       number of concurrent search workers (default 10)
  -d, --domains stringArray   domains to search for
  -f, --files stringArray     dataset files
  -h, --help                  help for fdnssearch
      --show-type             show record type for search results
      --show-value            show record value for search results
  -t, --types stringArray     record types to search for (default [a])
```

### Remote Datasets

TBD

### Local Datasets

TBD