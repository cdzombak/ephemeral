#!/usr/bin/env bash

set -eu

aws lambda update-function-code --function-name ephemeral-$EPHEMERAL_INSTANCE_NAME --zip-file fileb://dist/ephemeral.zip
