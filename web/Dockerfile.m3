# Build stage (same multi-page Vite build)
FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY src/ ./src/
COPY index.html m3.html tsconfig.json vite.config.ts ./
RUN npm run build

# Serve stage — Material 3 entry (m3.html)
FROM nginxinc/nginx-unprivileged:alpine
COPY nginx.m3.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
USER 101:101
EXPOSE 8080
