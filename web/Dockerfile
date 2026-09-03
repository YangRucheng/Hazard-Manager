# 前端镜像：多阶段构建，Node 打包 + Nginx 托管静态产物并反代 /api
FROM node:24-alpine AS builder

WORKDIR /build

# pnpm 版本与仓库一致
RUN corepack enable && corepack prepare pnpm@11.7.0 --activate

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]