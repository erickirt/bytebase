package store

import (
	"fmt"
	"net"
	"net/url"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	dbdriver "github.com/bytebase/bytebase/backend/plugin/db"
	storepb "github.com/bytebase/bytebase/proto/generated-go/store"
)

// GetEmbeddedConnectionConfig gets the embedded connection config.
func GetEmbeddedConnectionConfig(datastorePort int, pgUser string) dbdriver.ConnectionConfig {
	// Even when Postgres opens Unix domain socket only for connection, it still requires a port as ID to differentiate different Postgres instances.
	// For embedded database, the database name is the same as the user name.
	return dbdriver.ConnectionConfig{
		DataSource: &storepb.DataSource{
			Username: pgUser,
			Password: "",
			Host:     common.GetPostgresSocketDir(),
			Port:     fmt.Sprintf("%d", datastorePort),
			Database: pgUser,
		},
		ConnectionContext: dbdriver.ConnectionContext{
			DatabaseName: pgUser,
		},
		MaximumSQLResultSize: common.DefaultMaximumSQLResultSize,
	}
}

// GetConnectionConfig gets connection config from pgURL.
func GetConnectionConfig(pgURL string) (dbdriver.ConnectionConfig, error) {
	u, err := url.Parse(pgURL)
	if err != nil {
		return dbdriver.ConnectionConfig{}, err
	}
	q := u.Query()

	// Though the official libpq adopts postgresql:// (https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING)
	// Several popular services such as render.com, supabase use postgres://.
	// So we allow both schemes. The underlying pgx driver also supports both format.
	if u.Scheme != "postgresql" && u.Scheme != "postgres" {
		return dbdriver.ConnectionConfig{}, errors.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	connCfg := dbdriver.ConnectionConfig{
		DataSource:           &storepb.DataSource{},
		MaximumSQLResultSize: common.DefaultMaximumSQLResultSize,
	}

	if u.User != nil {
		connCfg.DataSource.Username = u.User.Username()
		connCfg.Password, _ = u.User.Password()
	}
	if connCfg.DataSource.Username == "" {
		return dbdriver.ConnectionConfig{}, errors.Errorf("missing user in the --pg connection string")
	}
	if host, port, err := net.SplitHostPort(u.Host); err != nil {
		connCfg.DataSource.Host = u.Host
	} else {
		// There is a hack. PostgreSQL document(https://www.postgresql.org/docs/14/libpq-connect.html)
		// specifies that a Unix-domain socket connection is chosen if the host part is either empty or **looks like an absolute path name**.
		// But url.Parse() does not meet this standard, for example:
		// url.Parse("postgresql://bbexternal@/tmp:3456/postgres"), it will consider `tmp:3456/postgres` as `path`,
		// and we use path as dbasename(same as PostgreSQL document) so that we get a wrong dbname.
		// So we put the socket path in the `host` key in the query,
		// note that in order to comply with the Postgresql document we are not using the `socket` key with obvious semantics.
		// To give a correct example: postgresql://bbexternal@:3456/postgres?host=/tmp
		hostInQuery := q.Get("host")
		if hostInQuery != "" && host != "" {
			// In this case, it is impossible to decide whether to use socket or tcp.
			return dbdriver.ConnectionConfig{}, errors.Errorf("please only using socket or host instead of both")
		}
		connCfg.DataSource.Host = host
		if hostInQuery != "" {
			connCfg.DataSource.Host = hostInQuery
		}
		connCfg.DataSource.Port = port
	}
	if connCfg.DataSource.Port == "" {
		connCfg.DataSource.Port = "5432"
	}
	if u.Path == "" {
		return dbdriver.ConnectionConfig{}, errors.Errorf("missing database in the --pg connection string")
	}
	connCfg.ConnectionContext.DatabaseName = u.Path[1:]
	connCfg.DataSource.SslCa = q.Get("sslrootcert")
	connCfg.DataSource.SslKey = q.Get("sslkey")
	connCfg.DataSource.SslCert = q.Get("sslcert")
	return connCfg, nil
}
