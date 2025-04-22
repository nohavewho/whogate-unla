FROM node:20.18.0 AS builder

WORKDIR /app/web

COPY web/package*.json ./

RUN npm install

COPY web/ .

RUN npm run build

FROM nginx:1.27.5-bookworm

COPY --from=builder /app/web/dist /usr/share/nginx/html

COPY deploy/docker/multi/nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"] 