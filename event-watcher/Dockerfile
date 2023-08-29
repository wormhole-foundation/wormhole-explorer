# It needs all dependencies (dev and prod)
# It's in charge of creating the dist folder

FROM node:18-alpine AS builder

WORKDIR /app

COPY . .

RUN npm install \
&& npm run build


# It needs only the production dependencies
# It's in charge of creating the final image

FROM node:18-alpine AS runner

WORKDIR /app

COPY --from=builder ./app/dist ./dist
COPY package.json .

RUN npm install --omit=dev

CMD [ "npm", "run", "start" ]