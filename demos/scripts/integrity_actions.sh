#!/usr/bin/env bash

# Script to run *within* the asciinema container to automate the integrity examples recordings

# Remove any existing checksums
echo "integrity_actions: removing any existing checksums"
integrity -dxrq /home/photos

echo "integrity_actions: set the prompt"

export TERM=xterm-256color
export PROMPT_COMMAND=''
export PS1='\[\e[36m\][\W]\[\e(B\e[m\]\n$ '

for file in demos/scripts/actions/*.sh; do
 if [ -f "$file" ]; then
    filename="${file%.*}"
    filename=$(basename -s .sh "$file")
    example="$filename"
    echo "integrity_examples: running ${example}"
    rm -f "demos/output/${example}.cast"
    asciinema-automation -dt 0 -sd 0 -aa "-c bash" "demos/scripts/actions/${example}.sh" "demos/output/${example}.cast"
    #agg "demos/scripts/actions/${example}.cast" "demos/scripts/actions/${example}.gif"
  fi
done