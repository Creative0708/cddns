# ~~Cloudflare~~ Colin's Dynamic DNS (CDDNS)

A sort of sub-API wrapper providing dynamic DNS for the [Cloudflare API](https://developers.cloudflare.com/api/), because Cloudflare isn't fine-grained enough.

Each CDDNS client has an API key, which they use to update IP address for one (!) subdomain by contacting the CDDNS server, which then updates the DNS record through Cloudflare. Basically, use this if you're e.g. hosting dynamic DNS for someone else on one of your domains and don't want to give them an API key to your entire domain.

Currently only Linux x86_64/aarch64 is supported. The "Run automatically on startup" installer only works for systemd-based distros.

# Usage

## Client

Download and run `cddns-client` for your platform. Follow the TUI options.

```
$ ./cddns-client
no config file found. entering setup
server URL: https://example.com/cddns
API key: CDDNS_ABCDEFGHIJKLMNOPQRSTUVWXYZ
? run automatically?:
  ▸ on reboot
    hourly
    daily
    weekly
    don't run automatically
```

If you don't have it run automatically, the client will periodically have to run the `cddns-client` binary (without options) to set the new IP address.

## Server

This is the complicated part.

Set `CLOUDFLARE_API_TOKEN` as an environment variable to get `cddns-server` to run. You can put this behind a reverse proxy and/or a URL prefix, just make sure it's accessible somewhere publicly. (Preferably with TLS.)

I don't have automatic setup for the server yet. I use this setup:

`/etc/systemd/system/cddns.service`:

```
[Unit]
Description=CDDNS (Colin's Dynamic DNS)
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=5
User=cddns
ExecStart=/var/lib/cddns/start.sh

[Install]
WantedBy=multi-user.target
```

`/var/lib/cddns/start.sh`:

```sh
#!/bin/sh
cd "$(dirname "$(realpath "$0")")"

export CLOUDFLARE_API_TOKEN="<api token>"
./cddns-server "$@"
```

You can then run `start.sh -new` to select a zone and record from the domains accessible to the API key.

# License

[MIT Licensed](/LICENSE). Made in a day. Have fun
