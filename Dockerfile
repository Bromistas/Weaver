FROM golang:1.21-alpine

WORKDIR /app

COPY src/go.mod ./
RUN go mod download

COPY src/common ./common
ARG APP_PATH
COPY $APP_PATH ./$APP_PATH

RUN go build -o /weaver-$APP_PATH

EXPOSE 8080

CMD [ "/weaver-$APP_PATH" ]