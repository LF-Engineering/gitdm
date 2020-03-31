#!/bin/bash
DA_API_URL=`cat secrets/DA_API_URL.secret` GITDM_GIT_REPO=`cat secrets/GITDM_GIT_REPO.secret` GITDM_GIT_USER=`cat secrets/GITDM_GIT_USER.secret` GITDM_GIT_OAUTH=`cat secrets/GITDM_GIT_OAUTH.secret` ./gitdm-sync
