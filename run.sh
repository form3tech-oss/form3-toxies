#!/bin/bash
go build -o toxi ./cmd/server/server.go
docker run -d -e POSTGRES_PASSWORD=postgres -p "5432:5432" postgres
go test ./... -v
# psql -H localhost -p 4321
