#!/bin/bash
# DRY=1 - dry run mode
# NS=gitdm - set namespace name, default gitdm
helm=helm
denv=test
if [ -z "$1" ]
then
  echo "$0: you should env: test, prod, using default helm"
else
  helm="${1}h.sh"
  denv="${1}"
fi
if [ -z "$NS" ]
then
  NS=gitdm
fi
if [ -z "$DRY" ]
then
  $helm install "${NS}-namespace" ./gitdm --set "namespace=$NS,skipSecrets=1,skipSync=1"
  change_namespace.sh $1 "$NS"
  $helm install "$NS" ./gitdm --set "namespace=$NS,deployEnv=$denv,skipNamespace=1"
  change_namespace.sh $1 default
else
  echo "Dry run mode"
  change_namespace.sh $1 "$NS"
  $helm install --debug --dry-run --generate-name ./gitdm --set "namespace=$NS,deployEnv=$denv,dryRun=1"
  change_namespace.sh $1 default
fi
