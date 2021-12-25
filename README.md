# fdnssearch

![Build Status](https://github.com/nscuro/fdnssearch/workflows/Continuous%20Integration/badge.svg?branch=master)

*Swiftly search [FDNS](ttps://github.com/rapid7/sonar/wiki/Forward-DNS) datasets from Rapid7 Open Data*

**Disclaimer**: You can do most of what *fdnssearch* does with [`bash`, `curl`, `pigz` and `jq`](https://github.com/rapid7/sonar/wiki/Analyzing-Datasets). This is nothing revolutionary. *fdnssearch* simply offers a [nicer UX](#usage) and some [QoL features](#interoparability). If you prefer a server-based solution, take a look at the [amazing Crobat project](https://github.com/Cgboal/SonarSearch).

## Installation

`go install github.com/nscuro/fdnssearch/cmd/fdnssearch@latest`

Alternatively, clone this repo and run `make install`. Make sure `$GOPATH/bin` is in your `$PATH`.

*fdnssearch* requires Go >= 1.17

Prebuilt binaries are available [as well](https://github.com/nscuro/fdnssearch/releases/).

### Docker

Clone this repository, `cd` into it and run `make docker`.  
The image can then be used as follows: `docker -it --rm nscuro/fdnssearch -h`

## Usage

```                                                  
Usage:
  fdnssearch [flags]

Flags:
      --amass-config string    Amass config to load domains from
  -a, --any                    Additionally search ANY dataset (ignored when -f is set)
      --any-only               Only search ANY dataset (ignored when -f is set)
  -d, --domains stringArray    Domains to search for
  -e, --excludes stringArray   Domains to exclude from search
  -f, --files stringArray      Dataset files
  -h, --help                   help for fdnssearch
  -o, --output string          Output file
      --plain                  Disable colored output
  -q, --quiet                  Only print results, no errors or log messages
      --timeout int            Timeout in seconds
  -t, --types stringArray      Record types to search for (a, aaaa, cname, txt, mx) (default [a])
  -v, --version                Show version
```

Errors and log messages are written to `STDERR`, search results to `STDOUT`. This allows for easy piping without the need to use `--quiet`. 
When piping results to other commands, make sure to disable colored output with `--plain`.

### Examples

Searching for `A` and `CNAME` records of subdomains of `example.de` and `example.com`:

```bash
$ fdnssearch -d example.de -d example.com -t a -t cname
```

Searching for `AAAA` and `TXT` records of subdomains of `example.com`, disabling colored output and writing results to `results.txt`:

```bash
$ fdnssearch -d example.com -t aaaa -t txt --plain | tee results.txt
```

Searching for `A` records of subdomains of `example.com`, excluding `(*.)acme.example.com` and writing results to `results.txt`:

```bash
$ fdnssearch -d example.com -e acme.example.com -o results.txt
```

### Remote Datasets

When no local dataset files are provided using `-f` / `--files`, *fdnssearch* will fetch the current datasets from Rapid7's website. It will search all datasets that match the record types provided with `-t` / `--types`. 

This requires a fairly good internet connection, but doesn't pollute your storage with huge files that get outdated quickly. The slower your connection, the fewer search workers are required.

Rapid7 provides a dataset with `ANY` records in addition to the specific datasets:

> Until early November 2017, all of these were for the 'ANY' record with a fallback A and AAAA request if neccessary. After that, the ANY study represents only the responses to ANY requests, and dedicated studies were created for the A, AAAA, CNAME and TXT record lookups with appropriately named files.

If you want your search to include this dataset as well, use the `--any` flag. Be aware that you **will** get a lot of duplicate results this way. Be sure to [deduplicate](#deduplication) your results. If you want to save time however, additionally pass the `--any-only` flag. *fdnssearch* will then exclusively search search the `ANY` dataset. While this may yield results faster, you may not get as many results as without `--any-only`.

### Local Datasets

It is possible to search local dataset files as well:

```bash
$ fdnssearch -f /path/to/datasets/2020-05-23-1590208726-fdns_a.json.gz -d example.com
```

### Performance

*fdnssearch* uses *klauspost*'s [`pgzip`](https://github.com/klauspost/pgzip) for performant decompression of the datasets.
Thanks to `pgzip`, the performance of *fdnssearch* is pretty much on par with the `pigz`, `grep` and `jq` approach:

```bash
$ time pigz -dc /path/to/datasets/2020-06-28-1593366733-fdns_cname.json.gz \
    | grep 'google\.com' \
    | jq '. | select(.name | endswith(".google.com")) | select(.type == "cname") | .name' \
    > /dev/null
pigz -dc /path/to/datasets/2020-06-28-1593366733-fdns_cname.json.gz  57.22s user 41.11s system 212% cpu 46.269 total
grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn,.idea,.tox}   18.78s user 6.91s system 55% cpu 46.268 total
jq  > /dev/null  2.59s user 0.07s system 5% cpu 46.268 total
```

```bash
$ time fdnssearch -d google.com -t cname --quiet \
    -f /path/to/datasets/2020-06-28-1593366733-fdns_cname.json.gz \
    > /dev/null
fdnssearch -d google.com -t cname -f  --quiet > /dev/null  64.58s user 0.85s system 144% cpu 45.266 total
```

This is with an [Intel i7 8700K](https://ark.intel.com/content/www/us/en/ark/products/126684/intel-core-i7-8700k-processor-12m-cache-up-to-4-70-ghz.html) and a [Samsung 970 EVO NVMe M.2 SSD](https://www.samsung.com/us/computing/memory-storage/solid-state-drives/ssd-970-evo-nvme-m2-500gb-mz-v7e500bw/) on Windows 10 in WSL 2.  
Your mileage may vary.

### Deduplication

*fdnssearch* will not perform deduplication in order to provide search results as quickly and efficiently as possible. 
Use tools like `uniq` or `sort` for this.

Given a file `results.txt` which only contains record names, deduplication can be achieved with:

```bash
$ sort --unique -o results.txt results.txt
```

### Interoparability

#### Amass

*fdnssearch* can parse target domains and exclusions from [Amass config files](https://github.com/OWASP/Amass/blob/master/examples/config.ini):

```bash
$ grep -A 3 "\[domains\]" amass.ini
[domains]
domain = example.com
domain = example.de
domain = example.fr

$ grep -A 1 "\[blacklisted\]" amass.ini
[blacklisted]
subdomain = acme.example.com

$ fdnssearch --amass-config amass.ini
```

This is equivalent to

```bash
$ fdnssearch -d example.com -d example.de -d example.fr -e acme.example.com
```
