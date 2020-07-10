# fdnssearch

![Build Status](https://github.com/nscuro/fdnssearch/workflows/Continuous%20Integration/badge.svg?branch=master)

## Installation

`GO111MODULE=on go get -v github.com/nscuro/fdnssearch/cmd/fdnssearch`

fdnssearch requires Go >= 1.14

## Usage

```
Usage:
  fdnssearch [flags]

Flags:
      --any                   additionally search ANY dataset (ignored when -f is set)
  -c, --concurrency int       number of concurrent search workers (default 10)
  -d, --domains stringArray   domains to search for
  -f, --files stringArray     dataset files
  -h, --help                  help for fdnssearch
      --no-ansi               disable ANSI output
      --show-type             show record type for search results
      --show-value            show record value for search results
      --silent                only print results, no errors or log messages
      --timeout int           timeout in seconds
  -t, --types stringArray     record types to search for (a, aaaa, cname, txt, mx) (default [a])
```

### Examples

Searching for `A` and `CNAME` records of subdomains of `example.de` and `example.com`, using `25` concurrent search workers:

```
fdnssearch -d example.de -d example.com -t a -t cname -c 25
```

Searching for `A` records of subdomains of `example.com`, disabling colored output and writing results to `results.txt`:

```
$ fdnssearch -d example.com --no-ansi | tee results.txt
```

### Remote Datasets

TBD

### Local Datasets

TBD

### Performance

TBD
