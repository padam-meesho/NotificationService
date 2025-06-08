package config

import (
	"strings"
	"sync"

	"github.com/gocql/gocql"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/utils"
	"github.com/scylladb/gocqlx/v2"
)

type ScyllaConfig struct {
	Hosts    []string
	Keyspace string
}

var (
	scyllaSessionOnce sync.Once
	scyllaDBSession   *ScyllaSession
)

type ScyllaSession struct {
	ScyllaSession *gocqlx.Session
}

func NewScyllaSession(appConfig *models.AppConfig) *gocqlx.Session {
	logger := utils.ComponentLogger("scylla")

	Hosts := strings.Split(appConfig.Scylla.Hosts, ",")
	logger.Info().
		Strs("hosts", Hosts).
		Str("keyspace", appConfig.Scylla.Keyspace).
		Msg("Initializing ScyllaDB connection")

	cluster := gocql.NewCluster(Hosts...)
	cluster.Keyspace = appConfig.Scylla.Keyspace
	cluster.Consistency = gocql.Quorum

	xSession, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		logger.Fatal().
			Err(err).
			Strs("hosts", Hosts).
			Str("keyspace", appConfig.Scylla.Keyspace).
			Msg("Failed to create ScyllaDB session")
	} else {
		logger.Info().
			Strs("hosts", Hosts).
			Str("keyspace", appConfig.Scylla.Keyspace).
			Msg("Successfully established ScyllaDB connection")
	}
	return &xSession
}

func InitScyllaSession(appConfig *models.AppConfig) *ScyllaSession {
	scyllaSessionOnce.Do(func() {
		scyllaDBSession = &ScyllaSession{
			ScyllaSession: NewScyllaSession(appConfig),
		}
	})
	return scyllaDBSession
}

func GetScyllaSession() *ScyllaSession {
	return scyllaDBSession
}
