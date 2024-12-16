# 构建前端dist
FROM node:20-alpine AS frontend
WORKDIR /frontend-build

COPY src/frontend .

RUN corepack enable && pnpm i --frozen-lockfile

RUN pnpm build

# 构建后端可执行文件
FROM golang:1.23-alpine AS backend
WORKDIR /backend-build

COPY src/backend .
COPY --from=frontend /frontend-build/dist /backend-build/static/dist

RUN CGO_ENABLED=1 go build -o nlip ./main.go

# 创建工作区并包含上述生成的文件
FROM alpine:latest AS monolithic
WORKDIR /usr/local/nlip

RUN apk add --no-cache tzdata
ENV TZ="Asia/Shanghai"

COPY --from=backend /backend-build/nlip /usr/local/nlip/

EXPOSE 3000

# 存储数据的目录，可以作为挂载点
RUN mkdir -p /var/opt/nlip
VOLUME /var/opt/nlip

ENV APP_ENV="production"
ENV SERVER_PORT="3000"

ENTRYPOINT ["./nlip"]