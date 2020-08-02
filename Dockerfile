FROM golang:alpine as go_builder

RUN pwd
WORKDIR /go/src/
COPY . /go/src/

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/godashboard main.go
RUN find . -name "godashboard"

# Build the React application
FROM node:alpine AS node_builder

COPY --from=go_builder /go/src/static /static
RUN find /go/src/static | grep -v "node_modules"
WORKDIR /static
RUN npm install
RUN npm run build

# Prepare final, minimal image
FROM alpine:latest
WORKDIR /app
RUN mkdir -p /app/static/

COPY --from=go_builder /go/src/bin/ /app
COPY --from=node_builder /static/build /app/static/build

# Install
ENV HOME /app

CMD ["./godashboard"]