#! /usr/bin/env bash
set -euo pipefail

# JQ as a github release
~/bin/jq --version || "jq command unsuccessful"

# a config file laid down
test -f ~/some-dir/some-config || "~/some-dir/some-config not present"
