FROM node:24-alpine3.24 AS builder

WORKDIR /app
RUN npm install --global pnpm@12.9.0
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build

FROM nginx:stable-alpine3.24

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
