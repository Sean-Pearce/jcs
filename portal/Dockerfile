FROM node:alpine3.11 as build-stage
WORKDIR /portal
COPY package*.json ./
RUN SASS_BINARY_SITE=https://npm.taobao.org/mirrors/node-sass/ npm install --registry=https://registry.npm.taobao.org
COPY ./ .
RUN npm run build:prod

FROM nginx:1.17.10-alpine as production-stage
RUN mkdir /portal
COPY --from=build-stage /portal/dist /portal
