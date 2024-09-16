#!/usr/bin/env bash

# Script to setup the docker container and files for integrity examples recordings

export DOCKER_CLI_HINTS=false

# Create the docker container and run in background
echo "Creating the docker container and run in background"
docker run -it --rm -d --name asciinema_auto \
--platform linux/amd64 \
-v "$(pwd)"/demos:/home/Alice/demos \
-v "$(pwd)"/bin/integrity_linux_arm64:/usr/bin/integrity \
pierremarchand/asciinema_playground

# Copy example binary files into the container
echo "Copying example binary files into the container"
docker cp "$(pwd)/pkg/integrity/testdata/imgs" asciinema_auto:/home/photos/

# Change the permissions of the files copied in
docker exec -it -u root asciinema_auto chmod -R 777 /home/photos

# Setup the prompt
#docker exec -it asciinema_auto bash -c "echo \"export TERM=xterm-256color; export PROMPT_COMMAND=''; export PS1='\n\[\e[36m\][\W]\[\e(B\e[m\]\n$ '\" >> /home/Alice/.bashrc"
docker exec -it asciinema_auto bash -c "echo \"export TERM=xterm-256color; export PROMPT_COMMAND=''; export PS1='\n\[\e[36m\]\$\[\e(B\e[m\] '\" >> /home/Alice/.bashrc"

# Run the examples script
echo "Running the examples script"
docker exec -it asciinema_auto demos/scripts/integrity_actions.sh

# Docker stop the container
echo "Stopping the container"
docker stop asciinema_auto