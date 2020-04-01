#!/bin/bash
if [ -z "${PR}" ]
then
  echo "$0: please provide PR number via PR=..."
  exit 1
fi
GITHUB_REF="refs/pull/${PR}/merge"
if [ -z "${SYNC_URL}" ]
then
  SYNC_URL='localhost:7070'
fi
(curl -s "${SYNC_URL}/pr/`git ls-remote 2>/dev/null | grep "${GITHUB_REF}" | awk '{ print $1 }'`" |& tee output.txt | grep 'CHECK_OK') || ( cat output.txt; exit 1)
