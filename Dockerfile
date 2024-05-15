FROM golang:1.21-alpine

WORKDIR /app
ARG ROLE

COPY src/$ROLE/go.mod ./
RUN go mod download

COPY src/$ROLE ./

RUN go build -o /weaver-$ROLE

EXPOSE 8080

# TODO: Dont hardcode the role
CMD [ "/weaver-storage" ]