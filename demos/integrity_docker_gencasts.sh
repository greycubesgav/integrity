#!/usr/bin/env bash

# Script to setup the docker container and files for integrity examples recordings
export DOCKER_CLI_HINTS=false

# Create the docker container and run in background
echo "Creating the docker container and run in background"

if ! docker run -it --rm -d --name integrity_demo \
  --platform linux/arm64 \
  -v "$(pwd)"/demos:/app/demos \
  greycubesgav/integrity-build
then
  echo "Failed to create the docker container"
  exit 1
fi

# Copy example binary files into the container
# These need to be copied into the container as we need to perform xattr operations on them
echo -e "\nCopying example binary files into the container"
docker cp "$(pwd)/pkg/integrity/testdata/imgs" integrity_demo:/home/photos/
docker cp "$(pwd)/pkg/integrity/testdata/data" integrity_demo:/home/data/
docker exec -it -u root integrity_demo chmod -R go+rwX /home/photos /home/data/

# Setup the prompt
#docker exec -it integrity_demo bash -c "echo \"export TERM=xterm-256color; export PROMPT_COMMAND=''; export PS1='\n\[\e[36m\][\W]\[\e(B\e[m\]\n$ '\" >> /home/Alice/.bashrc"
#docker exec -it integrity_demo bash -c "echo \"export TERM=xterm-256color; export PROMPT_COMMAND=''; export PS1='\n\[\e[36m\]\$\[\e(B\e[m\] '\" >> /home/Alice/.bashrc"

# Run the examples script
echo -e "\nRunning the examples script"
docker exec -it integrity_demo demos/scripts/integrity_actions.sh

# Docker stop the container
echo -e "\nStopping the container"
docker stop integrity_demo