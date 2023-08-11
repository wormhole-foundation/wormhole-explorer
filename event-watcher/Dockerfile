FROM node:18.17-alpine AS build
RUN apk add g++ make py3-pip

WORKDIR /app

COPY package*.json .
RUN npm install
COPY . .
RUN npm run build

############################
# STEP 2 build a small image
############################
FROM node:18.17-alpine AS run

COPY --from=build /app/dist ./dist

CMD ["node", "dist/src/index.js"]
