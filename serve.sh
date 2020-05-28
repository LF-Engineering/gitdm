#!/bin/bash
if [ -z "${JWT_TOKEN}" ]
then
  export JWT_TOKEN=`cat secrets/JWT_TOKEN.secret`
fi
if [ -z "${DA_API_URL}" ]
then
  export DA_API_URL=`cat secrets/DA_API_URL.secret`
fi
if [ -z "${GITDM_GITHUB_REPO}" ]
then
  export GITDM_GITHUB_REPO=`cat secrets/GITDM_GITHUB_REPO.secret`
fi
if [ -z "${GITDM_GITHUB_USER}" ]
then
  export GITDM_GITHUB_USER=`cat secrets/GITDM_GITHUB_USER.secret`
fi
if [ -z "${GITDM_GITHUB_OAUTH}" ]
then
  export GITDM_GITHUB_OAUTH=`cat secrets/GITDM_GITHUB_OAUTH.secret`
fi
if [ -z "${GITDM_GIT_USER}" ]
then
  export GITDM_GIT_USER=`cat secrets/GITDM_GIT_USER.secret`
fi
if [ -z "${GITDM_GIT_EMAIL}" ]
then
  export GITDM_GIT_EMAIL=`cat secrets/GITDM_GIT_EMAIL.secret`
fi
if [ -z "${AUTH0_URL}" ]
then
  export AUTH0_URL=`cat secrets/AUTH0_URL.secret`
fi
if [ -z "${AUTH0_AUDIENCE}" ]
then
  export AUTH0_AUDIENCE=`cat secrets/AUTH0_AUDIENCE.secret`
fi
if [ -z "${AUTH0_CLIENT_ID}" ]
then
  export AUTH0_CLIENT_ID=`cat secrets/AUTH0_CLIENT_ID.secret`
fi
if [ -z "${AUTH0_CLIENT_SECRET}" ]
then
  export AUTH0_CLIENT_SECRET=`cat secrets/AUTH0_CLIENT_SECRET.secret`
fi
./gitdm-sync
