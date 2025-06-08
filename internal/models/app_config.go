package models

type AppConfig struct {
	Kafka struct {
		BootStrapServers string // bootstrap.servers
		GroupId          string // "group.id"
		AutoOffsetReset  string // "auto.offset.reset"
	}
	Redis struct {
		Addr string // redis addr
		DB   int    // redis db
		Pwd  string // redis passwd
	}
	Scylla struct {
		Hosts    string // scylla hosts
		Keyspace string // scylla keyspace
	}
}
