package db

import (
	"context"
	"database/sql"
	"fmt"

	"community.threetenth.chatgpt/ent"
	"community.threetenth.chatgpt/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v4/stdlib"

	log "github.com/sirupsen/logrus"
)

var db *sql.DB
var client *ent.Client
var ctx context.Context

/*
OpenPostgreSQL is 打开并连接指定的 postgreSQL 数据库

su - postgres
psql

CREATE DATABASE database_name OWNER dbuser ENCODING 'UTF8';
GRANT ALL PRIVILEGES ON DATABASE exampledb TO dbuser;

最后创建和授权语句的结尾，一定一定一定要加分钟结束，否则不会生效，
没有任何错误提示，也没有任何提示。
因为是让你换行输入的。
*/
func OpenPostgreSQL(source string, debug bool) {
	var err error
	db, err = sql.Open("pgx", "postgresql://"+source)
	if err != nil {
		panic("open postgresql failed: " + err.Error())
	}

	drv := entsql.OpenDB(dialect.Postgres, db)

	opts := []ent.Option{
		ent.Driver(drv),
	}
	if debug {
		opts = append(opts, ent.Debug())
	}
	client = ent.NewClient(opts...)

	ctx = context.Background()
	err = client.Schema.Create(ctx,
		// migrate.WithGlobalUniqueID(true),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	if err != nil {
		log.Panicln("failed to create schema: ", err.Error())
	}
}

// WithTx best Practices, reusable function that runs callbacks in a transaction
func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v\n  rolling back transaction: %v", err.Error(), rerr.Error())
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
