#!/bin/sh
set -eu

: "${LANQIN_DATA_DIR:=/data}"
: "${LANQIN_DB_PATH:=/data/lanqin.db}"
: "${LANQIN_MAILDIR_ROOT:=/var/mail/vhosts}"
: "${LANQIN_ADDR:=127.0.0.1:8080}"

export LANQIN_DATA_DIR LANQIN_DB_PATH LANQIN_MAILDIR_ROOT LANQIN_ADDR
export LANQIN_DB_DRIVER=sqlite LANQIN_MAILSHARE_LITE=true
export LANQIN_SUBMISSION_ADDR="" LANQIN_SUBMISSION_TLS_ADDR=""

mkdir -p "${LANQIN_DATA_DIR}/attachments" "${LANQIN_MAILDIR_ROOT}"
nginx -t
nginx
# The API owns the container lifecycle. If the API exits, Docker restarts it.
# /readyz (through Nginx) detects either HTTP service being unavailable.
exec /usr/local/bin/lanqin-api
