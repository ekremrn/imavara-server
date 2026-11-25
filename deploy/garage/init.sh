#!/bin/sh
set -eu
node=$(/garage node id -q | cut -d@ -f1)
until /garage status >/dev/null 2>&1; do sleep 1; done
if ! /garage layout show | grep "^$node\|^$(echo "$node" | cut -c1-16).*local" >/dev/null; then
  /garage layout assign -z local -c 1G "$node"
  if /garage layout show | grep -q 'apply --version'; then
    version=$(/garage layout show | sed -n 's/.*apply --version \([0-9]*\).*/\1/p' | head -1)
    /garage layout apply --version "$version"
  fi
fi
if ! /garage key info "$S3_ACCESS_KEY" >/dev/null 2>&1; then
  /garage key import --yes -n imavara "$S3_ACCESS_KEY" "$S3_SECRET_KEY"
fi
for bucket in originals previews analysis snapshots; do
  /garage bucket info "$bucket" >/dev/null 2>&1 || /garage bucket create "$bucket"
  /garage bucket allow --read --write --key "$S3_ACCESS_KEY" "$bucket"
done
