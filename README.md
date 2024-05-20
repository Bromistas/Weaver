# Weaver
A distributed web scrapper

# Run rabbit mq
```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
```

# Build docker image
```bash
docker build --build-arg ROLE=router -t my-image .
```

# Run redis
```bash
docker run -d -p 8080:6379 --name redis1 redis
docker run -d -p 8070:6379 --name redis2 redis
```