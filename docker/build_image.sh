#!/bin/bash
if [ -z "${DOCKER_USER}" ]
then
  echo "$0: you need to set docker user via DOCKER_USER=username"
  exit 1
fi

make || exit 2
docker build -f ./docker/Dockerfile -t "${DOCKER_USER}/lf-gitdm-sync" . || exit 3
docker push "${DOCKER_USER}/lf-gitdm-sync" || exit 4
echo OK
