#!/bin/bash
#DESTROY the AKS cluster and everything on it

function  install_macosx_requirements () {

    printf "Installing terraform and tfenv on Mac OSX..."
    brew tap hashicorp/tap
    brew install tfenv
    tfenv install 0.14.11
    terraform -v
    tfenv use 0.14.11

}

function archive_data () {
  echo "TODO: backup important cluster data and results"
}

function delete_infrastructure () {
    terraform init
    terraform destroy
}

#install basic tooling + terraform version
install_macosx_requirements 
archive_data
delete_infrastructure
