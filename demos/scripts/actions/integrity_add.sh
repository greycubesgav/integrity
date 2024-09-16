echo "Check all photo files for existing checksums"
integrity /home/photos/*.*
echo "Add default sha1 checksum to all photo files"
integrity -a /home/photos/*.*
echo "List the default sha1 checksum for all photo files"
integrity -l /home/photos/*.*
echo "Check the file contents match the stored checksum"
integrity -c /home/photos/*.*
echo "Update the a photo file's contents and validate the checksum"
echo "Error data" >> /home/photos/_MG_5859.JPG
integrity -c /home/photos/_MG_5859.JPG
echo "Forcibly add a new checksum to the photo file"
integrity -af /home/photos/_MG_5859.JPG