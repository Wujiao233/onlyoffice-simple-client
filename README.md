# Onlyoffice Simple Client

A simple web client can upload&edit&save Office files based on Onlyoffice.

## Features

- Draw&Drop to upload and open office files
- Edit online based on onlyoffice documentserver
- Save each version after edit

## Deploy

1. first you need a Onlyoffice documentserver. See [Documentserver](https://hub.docker.com/r/onlyoffice/documentserver)
2. build the docker image 
```
git clone https://github.com/Wujiao233/onlyoffice-simple-client.git
cd onlyoffice-simple-client
docker build --no-cache --tag onlyoffice_simple_client:latest .
```
3. prepare data directory
4. create config file 
```
cp ./configs/config.json <path-to-your-data-directory>/configs/
vi <path-to-your-data-directory>/configs/config.json
```
5. set url and onlyoffice secret&host
6. run the docker
``` 
docker run --name=onlyoffice-simple-client \
        --env=TZ=Asia/Shanghai \
        --volume=<path-to-your-data-directory>/data:/app/data \
        --volume=/<path-to-your-data-directory>/configs:/app/configs \
        -p 8098:8098 onlyoffice_simple_client:latest
```

## Authentication

It's recommended to use `Authelia` to protect this service. just follow the [Documention](https://www.authelia.com/integration/proxies/nginx/), and add follow lines to your nginx conf file
```
location /files {
         proxy_pass $upstream;
}

location /callback {
         proxy_pass $upstream;
}
```

