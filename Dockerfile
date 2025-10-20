# Webpack build stage
FROM node:22-alpine AS webpack-builder
WORKDIR /app/web
COPY web/package.json ./
RUN npm install
COPY web/ ./
RUN npx webpack --mode=production

# Go build stage
FROM golang:latest AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
ARG GIT_TAG
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/ascii-arcade/farkle/config.Version=${GIT_TAG}" -a -installsuffix cgo -o ./bin/server .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=go-builder /app/bin/server /app/server
COPY --from=webpack-builder /app/web/dist /app/dist
CMD [ "./server" ]
