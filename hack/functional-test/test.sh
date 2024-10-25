#! /usr/bin/env bash
set -euo pipefail

# JQ as a github release
~/bin/jq --version || "jq command unsuccessful"

# a config file laid down
test -f ~/some-dir/some-config || "~/some-dir/some-config not present"

# htop should be installed
which htop

# nvm should exist and be checked out
test -d ~/github/nvm || "~/github/nvm does not exist"
cd ~/github/nvm
if [[ "$(git log -n 1 --oneline | cut -d ' ' -f 1)" != "179d450" ]]; then
    echo "nvm repo not checked out to correct commit"
    exit 1
fi

# neovim should be installed
/home/newuser/bin/neovim/bin/nvim --version

# Golang should be installed at 1.23.0
export PATH=$PATH:/usr/local/go/bin
if [[ "$(go version)" != "go version go1.23.0 linux/amd64" ]]; then
    echo "correct go version not installed"
    exit 1
fi

# dlv should be installed
~/go/bin/dlv -h

# kubectl should be installed
~/bin/kubectl --help || "kubectl unsuccessful"
