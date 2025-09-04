#!/bin/bash
# /usr/local/bin/check_stream_key.sh
STREAM_KEY=""
if [ "$1" != "$STREAM_KEY" ]; then
    exit 1  # deny publishing
fi
exit 0

sudo chmod +x /usr/local/bin/check_stream_key.sh
