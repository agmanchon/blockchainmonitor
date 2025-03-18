package postgresdb

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strconv"
)

type PostgresInfo struct {
	Host           string `mapstructure:"host" valid:"required~hosts is mandatory in SQL Connection config"`
	Port           int    `mapstructure:"port" valid:"required~port is mandatory in SQL Connection config"`
	Scheme         string `mapstructure:"scheme" valid:"required~scheme is mandatory in SQL Connection config"`
	Username       string `mapstructure:"username" json:"-" valid:"required~username is mandatory in SQL Connection config"`
	Password       string `mapstructure:"password" json:"-" valid:"required~password is mandatory in SQL Connection config"`
	SSLMode        string `mapstructure:"sslMode" valid:"required~sslMode is mandatory in SQL Connection config"`
	Database       string `mapstructure:"database" valid:"required~database is mandatory in SQL Connection config"`
	SchemaFilePath string `mapstructure:"schemaFilePath" valid:"required~database is mandatory in SQL Connection config"`
}

type PostgresConnectionFwOptions struct {
	Postgres *PostgresInfo `mapstructure:"postgres"`
}

type PostgresConnection interface {
	GetDB() *pgxpool.Pool
	ExecuteSchema() error
}
type Connection struct {
	db             *pgxpool.Pool
	schemaFilePath string
}

func (c Connection) ExecuteSchema() error {
	schema, err := os.ReadFile(c.schemaFilePath)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(context.Background(), string(schema))
	return err
}

func ConnectDB(options PostgresConnectionFwOptions) (PostgresConnection, error) {
	dsn := "postgres://" + options.Postgres.Username + ":" + options.Postgres.Password + "@" + options.Postgres.Host + ":" + strconv.Itoa(options.Postgres.Port) + "/" + options.Postgres.Database
	dbConnection, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Connection{db: dbConnection, schemaFilePath: options.Postgres.SchemaFilePath}, nil
}

func (c Connection) GetDB() *pgxpool.Pool {
	return c.db
}
