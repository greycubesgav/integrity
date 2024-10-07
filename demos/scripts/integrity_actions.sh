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

    echo "integrity_actions: removing any existing recordings"
    rm -f "demos/output/${example}.cast"

    echo "integrity_examples: running ${example} from ${file}"
    echo "demos/asciinema_automate.pl ${file}"
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


    #asciinema-automation -dt 0 -sd 0 -aa "-c bash" "demos/scripts/actions/${example}.sh" "demos/output/${example}.cast"
    #agg "demos/scripts/actions/${example}.cast" "demos/scripts/actions/${example}.gif"
  fi
done