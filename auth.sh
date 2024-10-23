#!/bin/sh
[[ "$(curl -s "https://www.duckdns.org/update?domains=${DUCKDNS_DOMAIN%.duckdns.org}&token=${DUCKDNS_TOKEN}&txt=${CERTBOT_VALIDATION}")" = "OK" ]]

# Repository: https://github.com/maksimstojkovic/docker-letsencrypt