#!/bin/bash

echo "[+] Performing passive amass enumeration..."
amass enum -passive -config amass.ini -nolocaldb -o domains.txt

echo "[+] Performing FDNS enumeration"
fdnssearch --amass-config amass.ini -t a -t aaaa -c 25 --no-ansi | tee -a domains.txt

echo "[+] Deduplicating results..."
sort --unique -o domains_dedupe.txt domains.txt

echo "[+] Cleaning up..."
rm results.txt && mv domains_dedupe.txt domains.txt

echo "[+] Resolving domains..."
massdns --resolvers $MASSDNS_PATH/lists/resolvers.txt -o J -w resolved_domains.json domains.txt
