# nextcloud-exporter [![Docker Build Status](https://img.shields.io/docker/build/xperimental/nextcloud-exporter.svg?style=flat-square)](https://hub.docker.com/r/xperimental/nextcloud-exporter/)

A [prometheus](https://prometheus.io) exporter for getting some metrics of a nextcloud server instance.

## Installation

If you have a working Go installation, getting the binary should be as simple as

```bash
go get github.com/xperimental/nextcloud-exporter
```

## Client credentials

To access the serverinfo API you will need the credentials of an admin user. It is recommended to create a separate user for that purpose.

## Usage

```plain
$ nextcloud-exporter --help
Usage of ./result/bin/nextcloud-exporter:
  -a, --addr string            Address to listen on for connections. (default ":9205")
  -p, --password string        Password for connecting to nextcloud.
      --password-file string   File containing the password for connecting to nextcloud.
  -t, --timeout duration       Timeout for getting server info document. (default 5s)
  -l, --url string             URL to nextcloud serverinfo page.
  -u, --username string        Username for connecting to nextcloud.
```

Some settings can also be specified through environment variables:

Name                     | Description
-------------------------|-------------------------------------
NEXTCLOUD_SERVERINFO_URL | URL to nextcloud serverinfo page
NEXTCLOUD_USERNAME       | Username for connecting to nextcloud
NEXTCLOUD_PASSWORD       | Password for connecting to nextcloud
NEXTCLOUD_PASSWORD_FILE  | File containing the password for connecting to nextcloud

Command line arguments take precedence over environment variables.

After starting the server will offer the metrics on the `/metrics` endpoint, which can be used as a target for prometheus.

The exporter will query the nextcloud server every time it is scraped by prometheus. If you want to reduce load on the nextcloud server you need to change the scrape interval accordingly:

```yml
scrape_configs:
  - job_name: 'nextcloud'
    scrape_interval: 90s
    static_configs:
      - targets: ['localhost:9205']
```

### Info URL

The exporter reads the metrics from the Nextcloud server using its "serverinfo" API. You can find the URL of this API in the administrator settings in the "Monitoring" section. It should look something like this:

```plain
https://example.com/ocs/v2.php/apps/serverinfo/api/v1/info
```

If you open this URL in a browser you should see an XML structure with the information that will be used by the exporter.
