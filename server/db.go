package main

import (
	"context"
	"sync"

	"adisuper94/turboguac/server/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool
var mutex = &sync.Mutex{}
var queries *generated.Queries

func getDBConn() *pgxpool.Pool {
	if pool == nil {
		mutex.Lock()
		if pool == nil {
			pool = createDBConnection(16)
		}
		mutex.Unlock()
	}

	return pool
}

func GetQueries() *generated.Queries {
	pool := getDBConn()
	if queries == nil {
		mutex.Lock()
		if queries == nil {
			queries = generated.New(pool)
		}
		mutex.Unlock()
	}

	return queries
}

func createDBConnection(connectionCount int32) *pgxpool.Pool {
	pgxConfig, err := pgxpool.ParseConfig("postgres://adisuper:password@localhost:5432/turbo?sslmode=disable")
	if err != nil {
		panic(err)
	}
	pgxConfig.MaxConns = connectionCount

	conn, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		panic(err)
	}
	return conn
}
