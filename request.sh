#!/bin/bash
if [ -z "${SYNC_URL}" ]
then
  SYNC_URL='localhost:7070'
fi
curl "${SYNC_URL}"
