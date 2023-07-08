# syntax=docker.io/docker/dockerfile:1.3@sha256:42399d4635eddd7a9b8a24be879d2f9a930d0ed040a61324cfdf59ef1357b3b2
FROM node:19.6.1-slim@sha256:a1ba21bf0c92931d02a8416f0a54daad66cb36a85d2b73af9d73b044f5f57cfc

# npm wants to clone random Git repositories - lovely.
# RUN apk add git python make build-base
# RUN apk update && apk add bash
RUN apt-get update && apt-get -y install \
  git python make curl netcat

RUN npm i typescript -g

USER 1000

RUN mkdir -p /home/node/app
RUN mkdir -p /home/node/.npm

WORKDIR /home/node/app

# Fix git ssh error
RUN git config --global url."https://".insteadOf ssh://

# Node
ENV NODE_EXTRA_CA_CERTS=/certs/cert.pem
ENV NODE_OPTIONS=--use-openssl-ca
# npm
RUN if [ -e /certs/cert.pem ]; then npm config set cafile /certs/cert.pem; fi
# git
RUN if [ -e /certs/cert.pem ]; then git config --global http.sslCAInfo /certs/cert.pem; fi

COPY --chown=node:node . .

RUN npm ci

run npm run build

CMD ["npm", "start"]




