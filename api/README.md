# SENDHOOKS-ENGINE-API

## Build docker image

```sh
cp .env.sample .env
cp config.sample.json config.json
# run apps
docker-compose up
```

## start locally sendhooks-engine-api

```sh
npm i && npm run dev
```

## test api

```sh
# post new hook
curl --location 'http://localhost:5001/api/send' \
--header 'Content-Type: application/json' \
--data '{
  "status": "success",
  "created": "12-02-2022"
}'

# fetch hooks
curl --location 'http://localhost:5002/api/sendhooks/v1/hooks'
```
