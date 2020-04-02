#!/bin/bash
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
./gitdm-sync
