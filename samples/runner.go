package samples

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/paveliak/go-workflows/backend"
	"github.com/paveliak/go-workflows/backend/mysql"
	"github.com/paveliak/go-workflows/backend/redis"
	"github.com/paveliak/go-workflows/backend/sqlite"
	"github.com/paveliak/go-workflows/diag"
	redisv8 "github.com/go-redis/redis/v8"
)

func GetBackend(name string, opt ...backend.BackendOption) backend.Backend {
	b := flag.String("backend", "redis", "backend to use: memory, sqlite, mysql, redis")
	flag.Parse()

	switch *b {
	case "memory":
		return sqlite.NewInMemoryBackend(opt...)

	case "sqlite":
		return sqlite.NewSqliteBackend(name+".sqlite", opt...)

	case "mysql":
		return mysql.NewMysqlBackend("localhost", 3306, "root", "root", name, opt...)

	case "redis":
		rclient := redisv8.NewUniversalClient(&redisv8.UniversalOptions{
			Addrs:        []string{"localhost:6379"},
			Username:     "",
			Password:     "RedisPassw0rd",
			DB:           0,
			WriteTimeout: time.Second * 30,
			ReadTimeout:  time.Second * 30,
		})

		rclient.FlushAll(context.Background()).Result()

		b, err := redis.NewRedisBackend(rclient, redis.WithBackendOptions(opt...))
		if err != nil {
			panic(err)
		}

		// Start diagnostic server under /diag
		m := http.NewServeMux()
		m.Handle("/diag/", http.StripPrefix("/diag", diag.NewServeMux(b)))
		go http.ListenAndServe(":3000", m)

		log.Println("Debug UI available at http://localhost:3000/diag")

		return b

	default:
		panic("unknown backend " + *b)
	}
}
