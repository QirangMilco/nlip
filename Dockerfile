# 构建前端dist
FROM node:20-alpine AS frontend
WORKDIR /frontend-build

COPY src/frontend .

RUN corepack enable && pnpm i --frozen-lockfile

RUN pnpm build

# 构建后端可执行文件
FROM golang:1.23-alpine AS backend
WORKDIR /backend-build

RUN apk add build-base

COPY src/backend .

RUN CGO_ENABLED=1 go build -o nlip ./main.go

# 创建工作区并包含上述生成的文件
FROM alpine:latest AS monolithic
WORKDIR /nlip

RUN apk add --no-cache tzdata
ENV TZ="Asia/Shanghai"

COPY --from=backend /backend-build/nlip /nlip/
COPY --from=frontend /frontend-build/dist /nlip/static/dist

EXPOSE 3000

ENV APP_ENV="production"
ENV SERVER_PORT="3000"

ENTRYPOINT ["./nlip"]