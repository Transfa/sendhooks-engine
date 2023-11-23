# SENDHOOKS-ENGINE-API

## start localy

```sh
npm i && npm run dev
```

## build docker image

```sh
# build sendhooks-engine-api image
docker build -t sendhooks-engine-api .

# run apps
docker-compose up
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
