#!/bin/bash
if [ -z "${DOCKER_USER}" ]
then
  echo "$0: you need to set docker user via DOCKER_USER=username"
  exit 1
fi
DA_API_URL=`cat secrets/DA_API_URL.secret`
GITDM_GITHUB_REPO=`cat secrets/GITDM_GITHUB_REPO.secret`
GITDM_GITHUB_USER=`cat secrets/GITDM_GITHUB_USER.secret`
GITDM_GITHUB_OAUTH=`cat secrets/GITDM_GITHUB_OAUTH.secret`
GITDM_GIT_USER=`cat secrets/GITDM_GIT_USER.secret`
GITDM_GIT_EMAIL=`cat secrets/GITDM_GIT_EMAIL.secret`
docker run -p 17070:7070 -e "DA_API_URL=${DA_API_URL}" -e "GITDM_GITHUB_REPO=${GITDM_GITHUB_REPO}" -e "GITDM_GITHUB_USER=${GITDM_GITHUB_USER}" -e "GITDM_GITHUB_OAUTH=${GITDM_GITHUB_OAUTH}" -e "GITDM_GIT_USER=${GITDM_GIT_USER}" -e "GITDM_GIT_EMAIL=${GITDM_GIT_EMAIL}" -it "${DOCKER_USER}/lf-gitdm-sync" "/usr/bin/gitdm-sync"
