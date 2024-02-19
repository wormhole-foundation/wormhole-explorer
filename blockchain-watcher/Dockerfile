# node:18.19.0-alpine3.18
FROM node@sha256:60d49c805812b23343b7be735784a4dad8f99ee4b97e5c7c939cb70f28534d4d AS builder
USER 1000

ARG GITHUB_TOKEN=null

RUN mkdir -p /home/node/app
RUN mkdir -p /home/node/.npm

WORKDIR /home/node/app

COPY --chown=node:node . .

RUN echo "//npm.pkg.github.com/:_authToken=${GITHUB_TOKEN}" > ./.npmrc && \
    npm ci && \
    rm -f ./.npmrc
RUN npm run build:ncc

# node:18.19.0-alpine3.18
FROM node@sha256:60d49c805812b23343b7be735784a4dad8f99ee4b97e5c7c939cb70f28534d4d AS runner

COPY --from=builder /home/node/app/config /home/node/app/config
COPY --from=builder /home/node/app/lib /home/node/app/lib

WORKDIR /home/node/app
COPY package.json .

CMD [ "npm", "run", "start:ncc" ]
