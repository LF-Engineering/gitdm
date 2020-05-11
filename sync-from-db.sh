#!/bin/bash
if [ -z "$1" ]
then
  echo "$0: please specify caller: github|ssaw"
  exit 1
fi
if [ -z "${SYNC_URL}" ]
then
  SYNC_URL='localhost:7070'
fi
(curl -s "${SYNC_URL}/sync-from-db/${1}" |& tee output.txt | grep 'SYNC_DB_OK') || ( cat output.txt; exit 1)
