#!/bin/bash
#
#made for mac osx. deal with it.

function destroy_kubernetes_cluster(){

  printf "destroy kubernetes ..."

  if [[ $target_cloud == "azure" ]]
  then
    destroy_aks_cluster
  fi

  if [[ $target_cloud == "aws" ]]
  then
    destroy_eks_cluster
  fi

}

function destroy_eks_cluster {
  printf "DESTROYING kubernetes cluster and everything on it!\n"
  mycwd=`pwd`
  cd kubernetes/eks-deploy/
  ./destroy.sh
  cd $mycwd
}

function destroy_aks_cluster {
  printf "DESTROYING kubernetes cluster and everything on it!\n"
  mycwd=`pwd`
  cd kubernetes/aks-deploy/
  ./destroy.sh
  cd $mycwd
}

#
destroy_kubernetes_cluster
