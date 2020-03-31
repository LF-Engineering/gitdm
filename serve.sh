#!/bin/bash
GITDM_GIT_REPO=`cat secrets/GITDM_GIT_REPO.secret` GITDM_GIT_USER=`cat secrets/GITDM_GIT_USER.secret` GITDM_GIT_OAUTH=`cat secrets/GITDM_GIT_OAUTH.secret` ./gitdm-sync
