#!/bin/bash
#
# Start script for insolvency-api
PORT="20188"

exec ./insolvency-api "-bind-addr=:${PORT}"
