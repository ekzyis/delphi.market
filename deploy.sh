#!/usr/bin/env bash

set -e

rsync -avh public/ vps:/var/www/delphi --delete --dry-run

echo
read -p "Continue deploy? [yn] " yn
echo
[ "$yn" == "y" ] && rsync -avh public/ vps:/var/www/delphi --delete
