FROM golang:1.19-alpine
RUN apk add \
  --no-cache \
  --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing \
  --repository http://dl-cdn.alpinelinux.org/alpine/edge/main \
	air
WORKDIR /home/snail
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . . 
EXPOSE 3002
CMD [ "air" ]
