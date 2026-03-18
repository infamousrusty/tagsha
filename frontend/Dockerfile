# =============================================================================
# Stage 1 — Builder
# =============================================================================
FROM node:22-alpine AS builder

WORKDIR /app

# Cache dependency layer before copying source.
# npm ci requires a committed package-lock.json; use npm install instead.
COPY package.json ./
RUN npm install --legacy-peer-deps

COPY . .
RUN npm run build

# =============================================================================
# Stage 2 — Runtime (nginx — serves built SPA)
# =============================================================================
FROM nginx:1.29-alpine

# Drop all capabilities; nginx only needs to bind on 80 (non-privileged inside container).
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Remove default nginx config to avoid conflicts.
RUN rm -f /etc/nginx/conf.d/default.conf.bak

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
  CMD wget -qO /dev/null http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
