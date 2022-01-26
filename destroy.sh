#!/bin/bash
#made for mac osx. deal with it.

function destroy_aks_cluster {
  printf "DESTROYING kubernetes cluster and everything on it!\n"
  mycwd=`pwd`
  cd kubernetes/aks-deploy/
  ./destroy.sh
  cd $mycwd
}

destroy_aks_cluster
