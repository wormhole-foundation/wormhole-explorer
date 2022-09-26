FROM node:16-alpine@sha256:f21f35732964a96306a84a8c4b5a829f6d3a0c5163237ff4b6b8b34f8d70064b

RUN mkdir -p /app
WORKDIR /app

COPY package.json package-lock.json ./
RUN --mount=type=cache,uid=1000,gid=1000,target=/home/node/.npm \
    npm ci
COPY . .
RUN npm run build
