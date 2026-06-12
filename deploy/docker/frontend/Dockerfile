# syntax=docker/dockerfile:1.7
FROM node:20-alpine AS build

WORKDIR /app

COPY frontend/package.json ./package.json
COPY frontend/vite.config.js ./vite.config.js
COPY frontend/index.html ./index.html
COPY frontend/src ./src

RUN npm install --legacy-peer-deps
RUN npm run build

FROM nginx:1.29-alpine

WORKDIR /usr/share/nginx/html

COPY deploy/docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist ./

EXPOSE 80
