# fdnssearch

![Build Status](https://github.com/nscuro/fdnssearch/workflows/Continuous%20Integration/badge.svg?branch=master)

*Swiftly search [FDNS](ttps://github.com/rapid7/sonar/wiki/Forward-DNS) datasets from Rapid7 Open Data*

**Disclaimer**: You can do most of what *fdnssearch* does with [`bash`, `curl`, `pigz`, `jq` and GNU `parallel`](https://github.com/rapid7/sonar/wiki/Analyzing-Datasets). This is nothing revolutionary. *fdnssearch* simply is nicer to use and 

## Installation

`GO111MODULE=on go get -v github.com/nscuro/fdnssearch/...`

Alternatively, clone this repo and run `make install`. Make sure `$GOPATH/bin` is in your `$PATH`.

*fdnssearch* requires Go >= 1.14

Prebuilt binaries are available [as well](https://github.com/nscuro/fdnssearch/releases/).

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
  -d, --domains stringArray    domains to search for
  -e, --excludes stringArray   domains to exclude from search
  -f, --files stringArray      dataset files
  -h, --help                   help for fdnssearch
      --plain                  disable colored output
  -q, --quiet                  only print results, no errors or log messages
      --timeout int            timeout in seconds
  -t, --types stringArray      record types to search for (a, aaaa, cname, txt, mx) (default [a])
```

Errors and log messages are written to `STDERR`, search results to `STDOUT`. This allows for easy piping without the need to use `--quiet`. When piping results to other commands, make sure to disable colored output with `--plain`.

### Examples

Searching for `A` and `CNAME` records of subdomains of `example.de` and `example.com`:

```bash
$ fdnssearch -d example.de -d example.com -t a -t cname
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
Each decompressed dataset entry immediately spawns a [goroutine](https://golangbot.com/goroutines/) ("*search worker*") that takes care of 
filtering and parsing. This means that the faster your source medium (internet connection, HDD or SSD), the more goroutines will run concurrently.
I/O really is the only limiting factor here, no matter where you load your datasets from.

For me, *fdnssearch* is even quite a bit faster than the `pigz`, `parallel` and `jq` approach:

```bash
$ time pigz -dc /path/to/datasets/2020-06-28-1593366733-fdns_cname.json.gz \
    | parallel --gnu --pipe "grep '\.google\.com'" \
    | parallel --gnu --pipe "jq '. | select(.name | endswith(\".google.com\")) | .name'" \
    > /dev/null
pigz -dc /path/to/datasets/2020-06-28-1593366733-fdns_cname.json.gz  62.84s user 41.95s system 113% cpu 1:32.02 total
parallel --gnu --pipe "grep '\.google\.com'"  185.31s user 78.92s system 287% cpu 1:32.02 total
parallel --gnu --pipe  > /dev/null  6.12s user 1.08s system 7% cpu 1:32.06 total

$ time fdnssearch -d google.com -t cname \
    -f /path/to/datasets/2020-06-28-1593366733-fdns_cname.json.gz \
    --quiet > /dev/null
fdnssearch -d google.com -t cname -f  --quiet > /dev/null  405.62s user 60.74s system 683% cpu 1:08.26 total
```

This is with an [Intel i7 8700K](https://ark.intel.com/content/www/us/en/ark/products/126684/intel-core-i7-8700k-processor-12m-cache-up-to-4-70-ghz.html) and a [Samsung 970 EVO NVMe M.2 SSD](https://www.samsung.com/us/computing/memory-storage/solid-state-drives/ssd-970-evo-nvme-m2-500gb-mz-v7e500bw/) on Windows 10 in WSL 2. Your mileage may vary.

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
