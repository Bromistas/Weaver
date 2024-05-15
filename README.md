# Weaver
A distributed web scrapper

# Run rabbit mq
```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
```

# Build docker image
```bash
docker build --build-arg APP_PATH=router -t my-image .
```