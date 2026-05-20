package cassandra

import (
	"fmt"
	"os"
	"time"

	"github.com/gocql/gocql"
)

func Connect(host string) *gocql.Session {
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "geoip"
	cluster.Consistency = gocql.LocalQuorum
	cluster.ConnectTimeout = 10 * time.Second

	for {
		session, err := cluster.CreateSession()
		if err == nil {
			return session
		}
		fmt.Fprintf(os.Stderr, "cassandra connect error: %v — retrying in 5s\n", err)
		time.Sleep(5 * time.Second)
	}
}
