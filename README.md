# Integrity
> Command line utility for storing, displaying and checking a file's checksum

## Features

* Supports Linux, FreeBSD and OSX
* Checksum data is stored in the file's extended attributes so can move with the file
* Multiple checksum algorithms available (defaults to sha1)
    * md4
    * md5
    * sha1
    * sha224
    * sha256
    * sha384
    * sha512
    * md5sha1
    * ripemd160
    * sha3 224
    * sha3 256
    * sha3 384
    * sha3 512
    * sha512 224
    * sha512 256
    * blake2s 256
    * blake2b 256
    * blake2b 384
    * blake2b 512

## Simple Usage examples

### Add checksum data to a file

```bash
integrity -a file.dat
file.dat : sha1 : added

```

### Display checksum data stored with file

```bash
integrity -l file.dat
```

### Validate the file still matches the stored checksum

```bash
integrity -c file.dat
```

## Advanced Usage Examples


### Add a sha256 checksum to a file
```bash
integrity --digest sha256 -a file.dat
```

Alternatively the Environment variable I_DIGEST may be used
```bash
export I_DIGEST="sha256"; integrity -a file.dat
```

### List all the checksums stored with a file

```bash
integrity -l -x file.dat
```




