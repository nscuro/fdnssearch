#!/bin/bash

echo "[+] Performing passive amass enumeration..."
amass enum -passive -config amass.ini -nolocaldb -o domains.txt

echo "[+] Performing FDNS enumeration..."
fdnssearch --amass-config amass.ini -t a -t aaaa --plain | tee -a domains.txt

echo "[+] Deduplicating results..."
sort --unique -o domains.txt domains.txt

echo "[+] Resolving domains..."
massdns --resolvers $MASSDNS_PATH/lists/resolvers.txt -o J -w domains_resolved.json domains.txt

# Probe for HTTP/S servers with httprobe
# ...
