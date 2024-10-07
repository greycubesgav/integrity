#!/usr/bin/env bash

# Script to run *within* the asciinema container to automate the integrity examples recordings

# Remove any existing checksums
echo "integrity_actions: removing any existing checksums"
integrity -dxrq /home/photos

for file in demos/scripts/actions/*.cmds; do
 if [ -f "$file" ]; then
    filename="${file%.*}"
    filename=$(basename -s .sh "$file")
    example="$filename"

    # Check if the .gif file is newer than the .cast file
    if [ -f "demos/output/${example}.gif" ] && [ -f "demos/scripts/actions/${example}" ] && [ "demos/output/${example}.gif" -nt "demos/scripts/actions/${example}" ]; then
      echo "integrity_actions: demos/output/${example}.gif is newer than demos/scripts/actions/${example}, skipping..."
      continue
    fi

    echo "integrity_actions: removing any existing recordings"
    rm -f "demos/output/${example}.cast"

    echo "integrity_actions: running [${example}] with demos/asciinema_automate.pl [${file}]"
    demos/asciinema_automate.pl "${file}"
    ret=$?
    if [ $ret -ne 0 ]; then
      echo "integrity_actions: ${example} failed to create casst, with return code $ret"
      exit $ret
    fi

    echo "Convert the recording to a gif"
    echo "agg demos/output/${example}.cast demos/output/${example}.gif"
    agg "demos/output/${example}.cast" "demos/output/${example}.gif"
    ret=$?
    if [ $ret -ne 0 ]; then
      echo "integrity_actions: ${example} agg failed to create gif, with return code $ret"
      exit $ret
    fi

  fi
done