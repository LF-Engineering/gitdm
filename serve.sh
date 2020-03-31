#!/bin/bash
DA_API_URL=`cat secrets/DA_API_URL.secret` GITDM_GITHUB_REPO=`cat secrets/GITDM_GITHUB_REPO.secret` GITDM_GITHUB_USER=`cat secrets/GITDM_GITHUB_USER.secret` GITDM_GITHUB_OAUTH=`cat secrets/GITDM_GITHUB_OAUTH.secret` ./gitdm-sync
