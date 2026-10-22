# syntax=docker/dockerfile:1
# Build context must be the repository root (npm workspaces + root lockfile).

FROM node:22-alpine AS build
WORKDIR /repo

# Install dependencies first for layer caching. Every workspace manifest named
# in the root package.json must be present for `npm ci`.
COPY package.json package-lock.json ./
COPY apps/web/package.json ./apps/web/
COPY apps/landing/package.json ./apps/landing/
RUN npm ci --workspace=apps/web --include-workspace-root=false

COPY apps/web ./apps/web

# The site is a static export, so the API URL is baked in at build time and is
# called from the user's browser (not from inside the container network).
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL NEXT_TELEMETRY_DISABLED=1
RUN npm run build --workspace=apps/web

FROM nginx:1.27-alpine
COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /repo/apps/web/out /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=15s --timeout=3s CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
