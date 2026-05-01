package config_test

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ichi-go/internal/infra/database"
)

func TestDatabasesMapUnmarshal(t *testing.T) {
	viper.Reset()
	viper.SetConfigType("yaml")

	yaml := `
database:
  default: mysql
  connections:
    mysql:
      driver: mysql
      host: db-mysql
      port: 3306
      name: app_db
      user: root
      password: secret
      max_idle_conns: 5
      max_open_conns: 25
      max_conn_life_time: 1800
      debug: false
    postgres:
      driver: postgres
      host: db-pg
      port: 5432
      name: queue_db
      user: pg_user
      password: pg_secret
      max_idle_conns: 3
      max_open_conns: 10
      max_conn_life_time: 900
      debug: false
`
	require.NoError(t, viper.ReadConfig(strings.NewReader(yaml)))

	var dbs map[string]database.Config
	require.NoError(t, viper.UnmarshalKey("database.connections", &dbs))

	assert.Len(t, dbs, 2)

	mysql := dbs["mysql"]
	assert.Equal(t, "mysql", mysql.Driver)
	assert.Equal(t, "db-mysql", mysql.Host)
	assert.Equal(t, 3306, mysql.Port)

	pg := dbs["postgres"]
	assert.Equal(t, "postgres", pg.Driver)
	assert.Equal(t, "db-pg", pg.Host)
	assert.Equal(t, 5432, pg.Port)

	assert.Equal(t, "mysql", viper.GetString("database.default"))
}
