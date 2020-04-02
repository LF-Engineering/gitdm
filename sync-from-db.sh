#!/bin/bash
if [ -z "${SYNC_URL}" ]
then
  SYNC_URL='localhost:7070'
fi
(curl -s "${SYNC_URL}/sync-from-db" |& tee output.txt | grep 'SYNC_DB_OK') || ( cat output.txt; exit 1)
