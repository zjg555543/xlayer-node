package main

import (
	"context"
	"fmt"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/jackc/pgx/v4/pgxpool"
	"reflect"
)

// NewSQLDB creates a new SQL DB
func NewSQLDB() (*pgxpool.Pool, error) {
	user := "state_user"
	password := "state_password"
	host := "127.0.0.1"
	port := "5432"
	name := "prover_db"

	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%d", user, password, host, port, name, 1000))
	if err != nil {
		log.Errorf("Unable to parse DB config: %v\n", err)
		panic(err)
	}

	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Errorf("Unable to connect to database: %v\n", err)
		panic(err)
	}
	return conn, nil
}

func main() {
	p, _ := NewSQLDB()
	tt := p.QueryRow(context.Background(), "select count(*) from state.program;")
	cnt := 0
	err := tt.Scan(cnt)
	fmt.Println("type()", err, reflect.TypeOf(tt))
	fmt.Println("tt", tt)
	fmt.Println("tt", cnt)
}
