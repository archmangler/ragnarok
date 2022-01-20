#!/bin/bash
#simple build script for producer image
#with registry push and versioning
#Usage: 
#   build push to remote reg: ./build.sh -n image_name -t tag -p yes
#E.g:
#./build.sh -n load-tester -t 0.0.1 -p true -d true 
#   build locally, no push to reg

registry_name="605125156525.dkr.ecr.ap-southeast-1.amazonaws.com"
repo_name="load-tester"
base_cmd="docker build --tag"

function build_image_local () {
  cmd="${base_cmd} ${image_name}:${image_tag} ."
  printf "running build command: $cmd\n"
  EXEC=$($cmd)
  echo $EXEC    
}

function tag_image () {

    cmd="docker tag ${image_name}:${image_tag} ${registry_name}/${repo_name}:${image_tag}"
    EXEC=$($cmd)
    echo $EXEC

}

function build_image_remote () {
  cmd="${base_cmd} ${image_name}:${image_tag} ."
  printf "running build command $cmd\n"
  EXEC=$($cmd)
  echo $EXEC    
}

function tag_image () {

    cmd="docker tag ${image_name}:${image_tag} ${registry_name}/${repo_name}:${image_tag}"
    EXEC=$($cmd)
    echo $EXEC

}

function push_image () {

    cmd="docker push ${registry_name}/${repo_name}:${image_tag}"
    printf "pushing docker image: ${cmd}"
    EXEC=$($cmd)
    echo $EXEC

}

while getopts n:t:p:d flag
   do
   case "${flag}" in
        n) image_name=${OPTARG};;
        t) image_tag=${OPTARG};;
        p) push_option=${OPTARG};;
#        d) debug=${OPTARG};;
   esac
done

printf "Building docker image with options: \n"
printf "Name: ${image_name}\n"
printf "Tag: ${image_tag}\n"
printf "Push: ${push_option}\n"

#NOTE:
repo_name=${image_name}

if [[ "${push_option}" == *"true"* ]]
then
  printf "building image ${image_name} with tag ${image_tag} and pushing to ${registry_name} repo ${repo_name}\n"
  build_image_remote
  tag_image
  push_image
else
  printf "building image ${image_name} with tag ${image_tag} locally\n"
  build_image_local
  tag_image
fi

#list the images
docker images
