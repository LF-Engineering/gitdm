#!/bin/bash
if [ -z "${DOCKER_USER}" ]
then
  echo "$0: you need to set docker user via DOCKER_USER=username"
  exit 1
fi
JWT_TOKEN=`cat token.secret`
DA_API_URL=`cat secrets/DA_API_URL.secret`
GITDM_GITHUB_REPO=`cat secrets/GITDM_GITHUB_REPO.secret`
GITDM_GITHUB_USER=`cat secrets/GITDM_GITHUB_USER.secret`
GITDM_GITHUB_OAUTH=`cat secrets/GITDM_GITHUB_OAUTH.secret`
GITDM_GIT_USER=`cat secrets/GITDM_GIT_USER.secret`
GITDM_GIT_EMAIL=`cat secrets/GITDM_GIT_EMAIL.secret`
AUTH0_URL=`cat secrets/AUTH0_URL.secret`
AUTH0_AUDIENCE=`cat secrets/AUTH0_AUDIENCE.secret`
AUTH0_CLIENT_ID=`cat secrets/AUTH0_CLIENT_ID.secret`
AUTH0_CLIENT_SECRET=`cat secrets/AUTH0_CLIENT_SECRET.secret`
docker run -p 17070:7070 -e "JWT_TOKEN=${JWT_TOKEN}" -e "AUTH0_URL=${AUTH0_URL}" -e "AUTH0_AUDIENCE=${AUTH0_AUDIENCE}" -e "AUTH0_CLIENT_ID=${AUTH0_CLIENT_ID}" -e "AUTH0_CLIENT_SECRET=${AUTH0_CLIENT_SECRET}" -e "DA_API_URL=${DA_API_URL}" -e "GITDM_GITHUB_REPO=${GITDM_GITHUB_REPO}" -e "GITDM_GITHUB_USER=${GITDM_GITHUB_USER}" -e "GITDM_GITHUB_OAUTH=${GITDM_GITHUB_OAUTH}" -e "GITDM_GIT_USER=${GITDM_GIT_USER}" -e "GITDM_GIT_EMAIL=${GITDM_GIT_EMAIL}" -it "${DOCKER_USER}/lf-gitdm-sync" "/usr/bin/gitdm-sync"
