# dns-duration
I'm currently having an issue on my local network where multiple times throughout the day, many devices on my network
have trouble performing DNS lookups. I have a pi-hole in my network, but it reports no issues, and my speedtests all
report properly as well. That led me to write this application that simply performs DNS lookups using multiple 
DNS servers, and report the time it takes to perform the lookup as a prometheus metric.

## Usage
This application supports the following arguments/environment variables:

| Argument | Environment Variable | Default Value | Help |
| -------- | -------------------- | ------------- | ---- |
| `--port`,`-p` | `PORT`               | `8080`        | The port number that metrics are served on |
| `--timeout`,`-t` | `TIMEOUT` | `5` | Amount of time, in seconds, to wait before failing the DNS failure |
| `--interval`,`-i` | `INTERVAL` | `15` | Amount of time, in minutes, to wait between lookups |
| `--name`,`-n` | `LOOKUP_HOSTNAME` | `google.com` | The hostname to use for performing DNS lookups |
| `--dns-server`,`-d` | `DNS_SERVER` | | The DNS server(s) to use for performing lookups, space separated |

When providing DNS servers, you can provide as many DNS servers as you want and a metric will be reported for each server. You should include the port that DNS is served on (usually `:53`. IE: `8.8.8.8:53`

## Exposed Metrics
Metrics are available on `/metrics`.

- `dns_lookup_speed`: Reports the speed of the DNS query 
` dns_errors_total`: Reports the number of times a DNS error occurred
