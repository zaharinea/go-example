# go-example

## Run service
```
make run
```

## Run linters
```
make lint
```

## Run service in docker
```
make docker-run
```

## Swagger docs
[http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html)


## API examples:

Create user
```
curl -X POST -H "Content-Type: application/json" -d '{"name": "user1"}' http://localhost:8000/api/users
{"id":"5fb5722853b2541a745bdc1c","name":"user1"}
```

Get users
```
curl -X GET http://localhost:8000/api/users
{"items":[{"id":"5fb5722853b2541a745bdc1c","name":"user1"}]}
```

Get user
```
curl -X GET http://localhost:8000/api/users/5fb5722853b2541a745bdc1c
{"id":"5fb5722853b2541a745bdc1c","name":"user1"}
```

Update user
```
curl -X PUT -H "Content-Type: application/json" -d '{"name": "user2"}' http://localhost:8000/api/users/5fb5722853b2541a745bdc1c
{"id":"5fb5722853b2541a745bdc1c","name":"user2"}
```

Delete user
```
curl -X DELETE http://localhost:8000/api/users/5fb5722853b2541a745bdc1c
```
