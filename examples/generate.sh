#!/bin/bash

set -e

# This script is more of an outline than a super reliable production-ready
# process, but it should serve as a reasonable starting point from which you
# can tweak things to your exact needs.
#
# Free for non-comercial uses
CITY=GeoLite2-City
ASN=GeoLite2-ASN
# Full versions
#CITY=GeoIP2-City
#ISP=GeoIP2-ISP

if [ -z "$MAXMIND_LICENSE_KEY" ]; then
    echo "MAXMIND_LICENSE_KEY env var required to download databases, even the free ones" >&2
    exit 1
fi

# A temporary place to download 
TMP="tmp"
mkdir -p $TMP

for edition_id in $CITY $ASN $ISP; do
  tar_gz="${TMP}/${edition_id}.tar.gz"
  if [ ! -e "$tar_gz" ]; then
    # We need to download it
    curl -sS --max-time 10 "https://download.maxmind.com/app/geoip_download?license_key=$MAXMIND_LICENSE_KEY&edition_id=$edition_id&suffix=tar.gz" > "$tar_gz"
    tar xvzf "$tar_gz" -C "$TMP"
    (cd $TMP && ln -s "${edition_id}"*"/${edition_id}.mmdb" "${edition_id}.mmdb")
  fi
done

OUT="out"
mkdir -p $OUT

if [ -z "$ISP" ]; then
  args="-asn-db ${TMP}/${ASN}.mmdb"
else
  args="-isp-db ${TMP}/${ISP}.mmdb"
fi
haproxy-mapper -outdir $OUT -all -city-db "${TMP}/GeoLite2-City.mmdb" $args

wc -l $OUT/*
