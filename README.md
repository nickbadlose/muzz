# muzz
Muzz tech test

## TODO
- makefile for docker up / restart db and clear data. Migrations should be ran in these too

## Postman

[Postman collection](https://www.postman.com/nickbadlose/workspace/muzz-api/collection/13188383-3d2cd57a-d0c4-43bb-ad64-0333f8a67deb?action=share&creator=13188383)

## Database Package

We use an SQL builder package, [upper/db](https://upper.io/v4/) to build SQL queries. Using an SQL builder forces 
paramterised queries, which helps protect against SQL injection. It also means we can freely migrate between any of the 
SQL variants supported by the lib without breaking changes. 

We create our own Database interface for clients as a facade for the unwanted complexities of the lib. This also allows
clients to decouple from the upper/db library, so if we wish to migrate to a different SQL builder, we can without 
breaking changes for clients of the package, admittedly with potential difficulties.

TODO
- package level doc - doc.go
- package level tests
- mock db
- state in README how unit testing is much easier with adapter pattern. We can use complicated sql mock on just repository methods and mock repositories for unit tests of service.

## Log Package

Wrapping the uber zap logger package with our own allows us to decorate logs with extra information, such as traceIDs
and decorate the package with extra methods, such as logger.MaybeError.

It also allows us to present the zap logging package as a global logger, which for logging I prefer global
loggers, as it means we can log in any deeply nested places with ease and without parameter drilling a logger. Or
without attaching multiple methods to the struct like so.

It also utilises the functional options pattern, which allows us to add configurations without breaking changes.

With composite logger:
```go
package mypackage

import "go.uber.org/zap"

type Server struct {
	logger *zap.Logger
}

func (s *Server) surfaceLevelFunc() {
	nestedFunc(s.logger)
} 

func nestedFunc(l *zap.Logger) {
	furtherNestedFunc(l)
}

func furtherNestedFunc(l *zap.Logger) {
	// we want to log in here
	l.Error(...)
}
```

With global logger:
```go
func furtherNestedFunc() {
	// we want to log in here
	logger.Error(...)
}
```

TODO
- package level doc - doc.go
- package level tests

