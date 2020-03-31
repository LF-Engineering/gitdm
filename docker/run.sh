#!/bin/bash
if [ -z "${DOCKER_USER}" ]
then
  echo "$0: you need to set docker user via DOCKER_USER=username"
  exit 1
fi
DA_API_URL=`cat secrets/DA_API_URL.secret`
GITDM_GIT_REPO=`cat secrets/GITDM_GIT_REPO.secret`
GITDM_GIT_USER=`cat secrets/GITDM_GIT_USER.secret`
GITDM_GIT_OAUTH=`cat secrets/GITDM_GIT_OAUTH.secret`
docker run -p 17070:7070 -e "GITDM_GIT_REPO=${GITDM_GIT_REPO}" -e "GITDM_GIT_USER=${GITDM_GIT_USER}" -e "GITDM_GIT_OAUTH=${GITDM_GIT_OAUTH}" -e "DA_API_URL=${DA_API_URL}" -it "${DOCKER_USER}/lf-gitdm-sync" "/usr/bin/gitdm-sync"
