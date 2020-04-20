FROM node:latest as build-stage
WORKDIR /portal
COPY package*.json ./
RUN npm install --registry=https://registry.npm.taobao.org
COPY ./ .
RUN npm run build:prod

FROM nginx as production-stage
RUN mkdir /portal
COPY --from=build-stage /portal/dist /portal
COPY nginx.conf /etc/nginx/nginx.conf