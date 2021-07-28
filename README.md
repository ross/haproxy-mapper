## HAProxy Mapper

Tools for generating [HAProxy
Maps](https://www.haproxy.com/blog/introduction-to-haproxy-maps/) from various
sources of data including [MaxMind](https://maxmind.com/),
[Spamhaus](https://www.spamhaus.org/), and various Cloud and CDN providers. The
resulting maps can be useful for many purposes: observability, legal
compliance, defending against DDoS attacks.

### Sources

* Providers
  * AWS
  * Azure - Hacky currently to avoid requiring auth
  * Cloudflare
  * Fastly
  * Google Cloud
  * Oracle
* MaxMind
  * ASN/ISP
  * City
* Spamhaus

If you have a source of data you'd like to see included [open an
issue](https://github.com/ross/haproxy-mapper/issues/new) or even better submit
a PR. If there's a publicly available JSON or TXT file with data for the source
please include info about it in the issue, doing so will make it much more
likely to be implemented.

### Mapfile Format

Map files have 1 value per line. The format is a range of IP addresses in [CIDR
notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation)
followed by a space and then an associated value for that range. The value
details vary by map, but consists of UTF-8 characters and punctuation and may
include spaces.  Generally haproxy-mapper uses `/` as a separator when a value
has multiple components. See the map file headers for more details.

```
1.0.0.0/24 OC/AU
1.0.1.0/24 AS/CN
1.0.2.0/23 AS/CN
1.0.4.0/22 OC/AU
1.0.8.0/21 AS/CN
1.0.16.0/20 AS/JP
1.0.32.0/19 AS/CN
1.0.64.0/24 AS/JP/Hiroshima/Hiroshima
1.0.65.0/25 AS/JP/Kanagawa/Sagamihara
1.0.65.128/25 AS/JP/Hiroshima/Hiroshima
```

## Using map files with HAProxy


### General map lookups

The following [HAProxy
configuration](https://cbonte.github.io/haproxy-dconv/2.4/configuration.html)
snippet will map the `src` to a location when one is available.

If you will be using the results of the map lookup for multiple purposes it
often makes sense to store the value in a variable. `txn` variables allow
access to the value across the full life-cycle of the request/connection once
their set. You can use `req` or `res` if your specific needs limit the times at
which you will need access to the value.

```
http-request set-var(txn.client_ip_location) src,map_ip(/etc/haproxy/maps/ip_to_location)
```

### Passing values to backend servers

Pass the previously mapped client IP location to the backend server as a
request header

```
http-request set-header x-client-ip-location %[var(txn.client_ip_location)]
```

### Making decisions based on lookups

Once you've looked up a value in a map it can be used to make decisions in
HAProxy via ACL's. For example the following would disallow requests that come
from IPs that have the country equal to Canada.

```
http-request set-var(txn.client_ip_location) src,map_ip(/etc/haproxy/maps/ip_to_location)
acl is_country_ca var(txn.client_ip_location) -i CA
use_backend 403_forbidden backend_banned_banhammer if is_country_ca
```

### Logging values

If you're going to be mapping IPs it often makes sense to log the results for
observability and debugging purposes. The details of HAProxy logging are beyond
the scope of this README, see
[ross/haproxied](https://github.com/ross/haproxied) for more information, but
the following snippet should provide the specific bits required to get a lookup
result into your logs. We'll first need to get the lookup value stored in a
variable.

```
http-request set-var(txn.client_ip_location) src,map_ip(/etc/haproxy/maps/ip_to_location)
```

We can then include the variable's value in our `log-format` to emit its value as part of our log line.

```
  log-format "backend_name=%b ... client_ip_location=%{+Q,+E}[var(txn.client_ip_location)] ..."
```

For map values that do not contain spaces or special characters the quoting and
escaping can be omitted, e.g ip_to_country which uses the 2-letter ISO codes.

```
  log-format "backend_name=%b ... client_ip_country=%[var(txn.client_ip_country)] ..."
```

### src vs X-Forwarded-For

The above example used
[`src`](https://cbonte.github.io/haproxy-dconv/2.4/configuration.html#src), but
in some situations that maybe the IP address of a proxy, either on the internet
or within your network, that will mask the true origin of the request. In many
such cases the original client IP address will be passed to the server using
the [`x-forwarded-for` header](https://en.wikipedia.org/wiki/X-Forwarded-For).

It is important that you "trust" the source of the header before you utilize it
as it is trivial for clients to add the header to confuse or obfuscate. This
can potentially be accomplished by having proxies under your control at the
edge of your network strip or clear any such headers they see. They can then
add a new x-forwarded-for header with the information as they see it.
Subsequent proxies in your system can then choose to leave x-forwarded-for
headers in place when `src` is on a trusted network and thus an internal server
under your control. See [ross/haproxied](https://github.com/ross/haproxied) for
more information and examples.

Assuming you trust the `x-forwarded-for` header you can make use of it during
mapping.

```
acl has_existing_x_forwarded_for req.hdr(x-forwarded-for) -m found
http-request set-header x-real-ip %[req.hdr(x-forwarded-for,1)] if has_existing_x_forwarded_for
http-request set-header x-real-ip %ci if !has_existing_x_forwarded_for
http-request set-header x-real-ip-location req.hdr(x-real-ip),map_ip(/etc/haproxy/maps/ip_to_location)
```

If you'll be using the values of `x-real-ip` and/or `x-real-ip-location`
elsewhere they can be set into variables.

```
http-request set-var(txn.real_ip) req.hdr(x-real-ip)
http-request set-var(txn.real_ip_location) req.hdr(x-real-ip-location)
```

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

## Related Links

* https://www.haproxy.com/blog/introduction-to-haproxy-maps/
* https://www.haproxy.com/documentation/hapee/latest/configuration/map-files/syntax/
* https://cbonte.github.io/haproxy-dconv/2.4/configuration.html
