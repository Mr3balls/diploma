FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci && chmod -R +x node_modules/.bin
COPY . .
# VITE_API_BASE_URL is injected at build time via Railway build variables
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# SPA fallback: redirect all routes to index.html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
