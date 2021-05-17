# denote

Self-destructive (one-time) notes.

Use POST form to set note and GET to get it. Sqlite for backend, note encryption with AES-GCM.

## Schema

POST form:
- value
- view_limit (default: 1)
- duration_limit (default: 24h, range: 1m..7*24h)

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
```

## Example

To run localhost:
```sh
$ URL="http://127.0.0.1:8000" denote
```

To run in docker:
```sh
$ docker-compose up -d
```

## Caveats

If you need frontend, setenv `ENABLE_STATIC=1`.

## Friends
- [potemkin](https://github.com/Termina1/potemkin)
- [shhh](https://github.com/smallwat3r/shhh)
- [tornote](https://github.com/osminogin/tornote)
