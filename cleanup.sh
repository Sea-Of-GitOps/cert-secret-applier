#!/bin/sh
[[ "$(curl -s "https://www.duckdns.org/update?domains=${DUCKDNS_DOMAIN%.duckdns.org}&token=${DUCKDNS_TOKEN}&txt=${CERTBOT_VALIDATION}&clear=true")" = "OK" ]]

# Repository: https://github.com/maksimstojkovic/docker-letsencrypt