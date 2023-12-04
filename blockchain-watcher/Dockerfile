FROM node:18.19.0-alpine3.18@sha256:c0a5f02df6e631b75ee3037bd4389ac1f91e591c5c1e30a0007a7d0babcd4cd3 as builder
USER 1000

RUN mkdir -p /home/node/app
RUN mkdir -p /home/node/.npm

WORKDIR /home/node/app

COPY --chown=node:node . .

RUN npm ci
RUN npm run build:ncc

FROM node:18.19.0-alpine3.18@sha256:c0a5f02df6e631b75ee3037bd4389ac1f91e591c5c1e30a0007a7d0babcd4cd3 as runner

COPY --from=builder /home/node/app/config /home/node/app/config
COPY --from=builder /home/node/app/lib /home/node/app/lib

WORKDIR /home/node/app
COPY package.json .

CMD [ "npm", "run", "start:ncc" ]




