## HAProxy Mapper

## Generating Maps

The shortcut `-all` will create all available maps. It requires both MaxMind
City and either ISP or ASN databases. If the ASN database is used ISP data is
not included in the generated maps.

```console
$ haproxy-mapper -outdir /tmp/mapper -all -isp-db GeoIP2-ISP.mmdb -city-db GeoIP2-City.mmdb
$ wc -l /tmp/mapper/*
  649426 /tmp/mapper/ip_to_asn
    4933 /tmp/mapper/ip_to_aws
   15868 /tmp/mapper/ip_to_azure
 3397950 /tmp/mapper/ip_to_city
      22 /tmp/mapper/ip_to_cloudflare
 3961657 /tmp/mapper/ip_to_continent
 3959544 /tmp/mapper/ip_to_country
      19 /tmp/mapper/ip_to_fastly
     433 /tmp/mapper/ip_to_google_cloud
  662153 /tmp/mapper/ip_to_isp
 3961657 /tmp/mapper/ip_to_location
     346 /tmp/mapper/ip_to_oracle
   21621 /tmp/mapper/ip_to_provider
    1132 /tmp/mapper/ip_to_spamhaus
 3399066 /tmp/mapper/ip_to_subdivisions
 20035827 total
```

### Full list of options

```console
$ haproxy-mapper
Usage of ./haproxy-mapper:
  -all
        Include All data
  -asn
        Include asn map, requires -asn-db or -isp-db
  -asn-db string
        MaxMind ASN database file
  -aws
        Include AWS data
  -azure
        Include Azure data
  -city
        Include city map, requires -city-db
  -city-db string
        MaxMind City database file
  -cloudflare
        Include Cloudflare data
  -continent
        Include continent map, requires -city-db
  -country
        Include country map, requires -city-db
  -fastly
        Include Fastly data
  -google-cloud
        Include Google Cloud data
  -ipv6
        Process IPv6 data, -ipv6=false to disable (default true)
  -isp
        Include isp map, requires -isp-db
  -isp-db string
        MaxMind ISP database file
  -location
        Include location map, requires -city-db
  -oracle
        Include Oracle data
  -outdir string
        Output directory
  -provider
        Include providers map
  -spamhaus
        Include spamhaus data
  -subdivisions
        Include subdivisions map, requires -city-db
```

## Example Script/Process

[examples/generate.sh](examples/generate.sh) is a starting point on which to
build a process for downloading MaxMind databases and building maps. It can be
run with the following. It uses the `-all` option, though as written relies on
the free databases and thus does not include the ISP map.

```
$ go build && PATH=$PATH:. ./examples/generate.sh
x GeoLite2-City_20210727/
x GeoLite2-City_20210727/README.txt
x GeoLite2-City_20210727/COPYRIGHT.txt
x GeoLite2-City_20210727/GeoLite2-City.mmdb
x GeoLite2-City_20210727/LICENSE.txt
x GeoLite2-ASN_20210727/
x GeoLite2-ASN_20210727/COPYRIGHT.txt
x GeoLite2-ASN_20210727/GeoLite2-ASN.mmdb
x GeoLite2-ASN_20210727/LICENSE.txt
Maps in out
  555291 out/ip_to_asn
    4936 out/ip_to_aws
   15868 out/ip_to_azure
 3451689 out/ip_to_city
      22 out/ip_to_cloudflare
 6117600 out/ip_to_continent
 6115499 out/ip_to_country
      19 out/ip_to_fastly
     433 out/ip_to_google_cloud
 6117600 out/ip_to_location
     346 out/ip_to_oracle
   21624 out/ip_to_provider
    1132 out/ip_to_spamhaus
 3766966 out/ip_to_subdivisions
 26169025 total
```

## Using map files with HAProxy

### General map lookups

### src vs X-Forwarded-For

### Passing values to backend servers

### Making decisions based on lookups

### Logging values
