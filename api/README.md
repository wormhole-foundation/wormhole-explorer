# API 

## How to build 

```bash
make build
```

## Config

You will need to set some env variables with the prefix `WORMSCAN` 

- WORMSCAN_DB_MONGO
- WORMSCAN_DB_NAME
- WORMSCAN_PORT 

for example: 

```bash
WORMSCAN_DB_URL=mongodb://localhost:27017/wormhole WORMSCAN_PORT=5555 ./api
```

## API Documentation

Documentation is automagically generated via swaggo using annotations on code
and placed inside `doc/` folder. 

To install swag tool run this 

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

To generate or update the doc run:

```bash
make doc
```