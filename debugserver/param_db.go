package debugserver

import (
	"database/sql"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// SqlDBStats is the state of sqldb
type SqlDBStats struct {
	// OpenConnections is the number of open connections to the database.
	OpenConnections int
	FreeConnections int
	UsedConnections int
	MaxIdle         int
	MaxOpen         int
}

func getSqlDBStatsReflect(db *sql.DB) SqlDBStats {

	v := reflect.ValueOf(*db)

	openConns := v.FieldByName("numOpen").Int()
	freeConns := v.FieldByName("freeConn").Len()
	maxIdle := v.FieldByName("maxIdle").Int()
	maxOpen := v.FieldByName("maxOpen").Int()

	return SqlDBStats{
		OpenConnections: int(openConns),
		FreeConnections: freeConns,
		MaxIdle:         int(maxIdle),
		MaxOpen:         int(maxOpen),
		UsedConnections: int(openConns) - freeConns,
	}
}

// AddParamSqlDB add a sql state
func AddParamSqlDB(name string, sqlDB *sql.DB) {
	AddParam(name, func() interface{} { return getSqlDBStatsReflect(sqlDB) })
}

type RedisStats struct {
	ActiveCount int
	IdleCount   int
	MaxIdle     int
	MaxActive   int
}

func AddParamRedis(name string, redisPool *redis.Pool) {
	AddParam(name, func() interface{} {
		poolStats := redisPool.Stats()
		return RedisStats{
			ActiveCount: poolStats.ActiveCount,
			IdleCount:   poolStats.IdleCount,
			MaxIdle:     redisPool.MaxIdle,
			MaxActive:   redisPool.MaxActive,
		}
	})
}
