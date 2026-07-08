FROM node:18-alpine AS builder
WORKDIR /app
RUN npm i -g pnpm
COPY . .
# Since it's a turborepo, we use turbo to prune and build
# For this scaffolding, assume standard build
RUN pnpm install
RUN pnpm turbo run build --filter=auth-service

FROM node:18-alpine AS runner
WORKDIR /app
COPY --from=builder /app/apps/auth-service/dist ./dist
COPY --from=builder /app/apps/auth-service/package.json ./
RUN npm i -g pnpm && pnpm install --prod
CMD ["node", "dist/main"]
