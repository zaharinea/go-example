# go-example

## Run service
```
cp .env.example .env
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


## Migrations
Install [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/database/mongodb)
```
brew install golang-migrate
```

Create new migration
```
make new-migration NAME="create_indexes"
```

Apply all migrations
```
make apply-migrations
```

Revert all migrations
```
revert-migrations
```

## API examples:

Create user
```
curl -X POST -H "Content-Type: application/json" -d '{"name": "user1"}' http://localhost:8000/api/users
{
    "id":"5fb5722853b2541a745bdc1c",
    "name":"user1", 
    "created_at":"2020-11-20T22:56:57.565Z",
    "updated_at":"2020-11-20T22:56:57.565Z"
}
```

Get users
```
curl -X GET http://localhost:8000/api/users
{
    "items":[
        {
            "id":"5fb5722853b2541a745bdc1c",
            "name":"user1",
            "created_at":"2020-11-20T22:56:57.565Z",
            "updated_at":"2020-11-20T22:56:57.565Z"
        }
    ]
}
```

Get user
```
curl -X GET http://localhost:8000/api/users/5fb5722853b2541a745bdc1c
{
    "id":"5fb5722853b2541a745bdc1c",
    "name":"user1", 
    "created_at":"2020-11-20T22:56:57.565Z",
    "updated_at":"2020-11-20T22:56:57.565Z"
}
```

Update user
```
curl -X PUT -H "Content-Type: application/json" -d '{"name": "user2"}' http://localhost:8000/api/users/5fb5722853b2541a745bdc1c
{
    "id":"5fb5722853b2541a745bdc1c",
    "name":"user1", 
    "created_at":"2020-11-20T22:56:57.565Z",
    "updated_at":"2020-11-20T22:58:02.686Z"
}
```

Delete user
```
curl -X DELETE http://localhost:8000/api/users/5fb5722853b2541a745bdc1c
```
