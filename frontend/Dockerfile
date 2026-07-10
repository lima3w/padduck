# syntax=docker/dockerfile:1
FROM node:22.22.3-alpine AS builder
WORKDIR /app
RUN npm install -g npm@11.16.0
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm,sharing=locked \
    npm ci --prefer-offline
COPY . .
RUN npm run build

FROM nginx:1.31.2-trixie
RUN apt-get update && apt-get upgrade -y --no-install-recommends && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD curl -fs http://127.0.0.1:3000/health | grep -q '"status":"ok"' || exit 1
CMD ["nginx", "-g", "daemon off;"]
