#!/bin/bash
BUCKET=$1
PREKEY=$2
HOST=`echo $HOSTNAME`
FULLPATH="s3://"$BUCKET"/"$PREKEY"/"$HOST"/
echo "$FULLPATH"
DIR="/etc/cipher"
if [[ -d "$DIR" ]]; then
  cd "$DIR"
  aws s3 cp ./ "$FULLPATH" --recursive
else
  echo "$DIR not found"
  exit(1)
fi
