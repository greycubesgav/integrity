
[![Test, Build & Package](https://github.com/greycubesgav/integrity/actions/workflows/package.yml/badge.svg)](https://github.com/greycubesgav/integrity/actions/workflows/package.yml)

# Integrity
> Command line utility for storing, displaying and checking a file's checksum

## Features

* Supports Linux, FreeBSD and OSX
* Checksum data is stored in the file's extended attributes so can move with the file
* Multiple checksum algorithms available (defaults to sha1)

| MD Functions | SHA Functions | SHA3 + SHA512 Functions  | Blake Functions | Misc Functions |
|--------------|---------------|--------------------------|-----------------|----------------|
| md5          | **[sha1]**          | sha3 224                 | blake2s 256     | phash          |
| md5sha1      | sha224        | sha3 256                 | blake2b 256     | ohash          |
|              | sha256  | sha3 384                 | blake2b 384     |
|              | sha384        | sha3 512                 | blake2b 512     |
|              | sha512        | sha512 224               |                 |
|              |               | sha512 256               |                 |

## Simple Usage examples

### Add checksum data to a file

```bash
$> integrity -a file.dat
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


### Add a blake2b_256 checksum to a file
```bash
integrity --digest blake2b_256 -a file.dat
```

Alternatively the Environment variable INTEGRITY_DIGEST may be used
```bash
export INTEGRITY_DIGEST="sha256"; integrity -a file.dat
```

### List all the checksums stored with a file

```bash
integrity -l -x file.dat
```




