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
  Usage of nextcloud-exporter:
    -a, --addr string          Address to listen on for connections. (default ":9205")
    -c, --config-file string   Path to YAML configuration file.
    -p, --password string      Password for connecting to Nextcloud.
    -t, --timeout duration     Timeout for getting server info document. (default 5s)
    -l, --url string           URL to Nextcloud serverinfo page.
    -u, --username string      Username for connecting to Nextcloud.
```

After starting the server will offer the metrics on the `/metrics` endpoint, which can be used as a target for prometheus.

### Configuration methods

There are three methods of configuring the nextcloud-exporter (higher methods take precedence over lower ones):

- Environment variables
- Configuration file
- Command-line parameters

#### Environment variables

All settings can also be specified through environment variables:

    Environment variable | Flag equivalent
-----------------------: | :--------------
NEXTCLOUD_LISTEN_ADDRESS | --addr
NEXTCLOUD_PASSWORD       | --password
NEXTCLOUD_TIMEOUT        | --timeout 
NEXTCLOUD_SERVERINFO_URL | --url
NEXTCLOUD_USERNAME       | --username

#### Configuration file

The `--config-file` option can be used to read the configuration options from a YAML file:

```yaml
listenAddress: ":9205"
password: "example"
timeout: "5s"
infoUrl: "https://example.com"
username: "example"
```

### Password file

Optionally the password can be read from a separate file instead of directly from the input methods above. This can be achieved by setting the password to the path of the password file prefixed with an "@":

```bash
$ nextcloud-exporter -c config.yml -p @/path/to/passwordfile
```

## Other information

### Info URL

The exporter reads the metrics from the Nextcloud server using its "serverinfo" API. You can find the URL of this API in the administrator settings in the "Monitoring" section. It should look something like this:

```plain
https://example.com/ocs/v2.php/apps/serverinfo/api/v1/info
```

When you do not specify a path on the `--url` parameter then the default path will be added automatically.

If you open this URL in a browser you should see an XML structure with the information that will be used by the exporter.

### Scrape configuration

The exporter will query the nextcloud server every time it is scraped by prometheus. If you want to reduce load on the nextcloud server you need to change the scrape interval accordingly:

```yml
scrape_configs:
  - job_name: 'nextcloud'
    scrape_interval: 90s
    static_configs:
      - targets: ['localhost:9205']
```
