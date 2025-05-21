#!/bin/bash
#
# Start script for insolvency-api 
PORT="${BIND_ADDR:=18101}"

if [[ ! -x ./insolvency-api ]]; then
  echo "ERROR: ./insolvency-api not found or not executable"
  exit 1
fi

exec ./insolvency-api "-bind-addr=:${PORT}"
