# denote

One-time notes. Use POST form to set note and GET query to get it. 
Sqlite3 for backend. Key derivation via Argon2id, note encryption with AES-GCM.

## Schema

POST form:
- value
- password (optional)
- view_limit (default: 1)
- duration_limit (default: 24h, min: 1m)

## Installation
```sh
$ make
$ make install
```
or
```sh
$ make docker
```

## Usage
```text
Usage of denote:
  -V	Print version and exit
  -h	Print help and exit
```

## Example

To run localhost:
```sh
$  URL_ORIGIN="http://127.0.0.1:8000" PASSWORD="qwertyuiopasdfgh" denote
```

To run in docker:
```sh
$ docker-compose up -d
```

## Friends
- [potemkin](https://github.com/Termina1/potemkin)
