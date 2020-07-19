# fdnssearch

![Build Status](https://github.com/nscuro/fdnssearch/workflows/Continuous%20Integration/badge.svg?branch=master)

**Disclaimer**: You can do most of what *fdnssearch* does with [`bash`, `curl`, `pigz`, `jq` and GNU `parallel`](https://github.com/rapid7/sonar/wiki/Forward-DNS).  
This is nothing revolutionary, I made this because I prefer simple commands over wonky shell scripts.

## Installation

`GO111MODULE=on go get -v github.com/nscuro/fdnssearch/cmd/fdnssearch`

Alternatively, clone this repo and run `make install`. Make sure `$GOPATH/bin` is in your `$PATH`.

*fdnssearch* requires Go >= 1.14

### Docker

Clone this repository, `cd` into it and run `make docker`.  
The image can then be used as follows: `docker -it --rm nscuro/fdnssearch -h`

## Usage

```
Usage:
  fdnssearch [flags]

Flags:
      --amass-config string    amass config to load domains from
  -a, --any                    additionally search ANY dataset (ignored when -f is set)
  -c, --concurrency int        number of concurrent search workers (default 10)
  -d, --domains stringArray    domains to search for
  -e, --excludes stringArray   domains to exclude from search
  -f, --files stringArray      dataset files
  -h, --help                   help for fdnssearch
      --plain                  disable colored output
  -q, --quiet                  only print results, no errors or log messages
      --show-type              show record type for search results
      --show-value             show record value for search results
      --timeout int            timeout in seconds
  -t, --types stringArray      record types to search for (a, aaaa, cname, txt, mx) (default [a])
```

Errors and log messages are written to `STDERR`, search results to `STDOUT`. This allows for easy piping without the need to use `--quiet`. When piping results to other commands, make sure to disable colored output with `--plain`.

### Examples

Searching for `A` and `CNAME` records of subdomains of `example.de` and `example.com`, using `25` concurrent search workers:

```bash
$ fdnssearch -d example.de -d example.com -t a -t cname -c 25
```

Searching for `AAAA` and `TXT` records of subdomains of `example.com`, disabling colored output and writing results to `results.txt`:

```bash
$ fdnssearch -d example.com -t aaaa -t txt --plain | tee results.txt
```

### Remote Datasets

When no local dataset files are provided using `-f` / `--files`, *fdnssearch* will fetch the current datasets from Rapid7's website. It will search all datasets that match the record types provided with `-t` / `--types`. 

This requires a fairly good internet connection, but doesn't pollute your storage with huge files that get outdated quickly. The slower your connection, the fewer search workers are required.

Rapid7 provides a dataset with `ANY` records in addition to the specific datasets:

> Until early November 2017, all of these were for the 'ANY' record with a fallback A and AAAA request if neccessary. After that, the ANY study represents only the responses to ANY requests, and dedicated studies were created for the A, AAAA, CNAME and TXT record lookups with appropriately named files.

If you want your search to include this dataset as well, use the `--any` flag. Be aware that you **will** get a lot of duplicate results this way. Be sure to [deduplicate](#deduplication) your results. 

### Local Datasets

It is possible to search local dataset files as well:

```bash
$ fdnssearch -f /path/to/datasets/2020-05-23-1590208726-fdns_a.json.gz -d example.com
```

### Performance

*fdnssearch* utilizes the `klauspost/pgzip` library for [performant gzip decompression](https://github.com/klauspost/pgzip#decompression-1).
Decompressed dataset entries are immediately submitted to a pool of [goroutines](https://golangbot.com/goroutines/) ("*search workers*") that take care of parsing and filtering. 
The size of this pool can be manipulated using the `-c` / `--concurrency` flag. The faster the source medium (internet connection, HDD, SSD), the bigger the pool should be for optimal performance.

### Deduplication

*fdnssearch* will not perform deduplication in order to provide search results as quick and efficient as possible. Use tools like `uniq` or `sort` for this.

Given a file `results.txt` which only contains record names, deduplication can be achieved with:

```bash
$ sort --unique -o results.txt results.txt
```

### Interoparability

#### Amass

*fdnssearch* can parse target domains and exclusions from [Amass config files](https://github.com/OWASP/Amass/blob/master/examples/config.ini):

```bash
$ grep -C 5 "\[domains\]" amass.ini | tail -6
[domains]
domain = example.com
domain = example.de
domain = example.fr

$ grep -C 1 "\[blacklisted\]" amass.ini | tail -2
[blacklisted]
subdomain = acme.example.com

$ fdnssearch --amass-config amass.ini
```

This is equivalent to

```bash
$ fdnssearch -d example.com -d example.de -d example.fr -e acme.example.com
```
