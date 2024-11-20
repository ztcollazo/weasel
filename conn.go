package weasel

import (
	"fmt"
	"net/url"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// Connection holds all of the connection information, to be used in the models.
// It also contains the configured raw query builder and DB API, which both are not
// recommended to directly be used, but are useful if you need a complex query that is
// not officially supported.
// The query builder type comes from Squirrel, and the DB type is *sqlx.DB.
type Connection struct {
	Builder sq.StatementBuilderType
	DB      *sqlx.DB
	driver  string
}

// Opts represents a generalized connection options structure for the Connect function.
// It is formatted automatically into a DSN based on the provided driver.
type Opts struct {
	// Common options
	User         string // Username for authentication
	Password     string // Password for authentication
	Host         string // Hostname or IP
	Port         int    // Port number
	Database     string // Database name
	Charset      string // Character set (e.g., utf8)
	TLS          string // TLS mode (e.g., "true", "false", "skip-verify")
	Timeout      int    // Connection timeout in seconds
	ReadTimeout  int    // Read timeout in seconds
	WriteTimeout int    // Write timeout in seconds

	// MySQL-specific options
	ParseTime            bool   // Parse time values into time.Time
	Collation            string // Collation to use (e.g., "utf8_general_ci")
	AllowNativePasswords bool   // Allow native password authentication
	MultiStatements      bool   // Allow multiple statements in one query

	// PostgreSQL-specific options
	SSLMode                 string // SSL mode (e.g., "disable", "require")
	SearchPath              string // Schema search path
	ApplicationName         string // Name of the application
	FallbackApplicationName string // Fallback app name
	ConnectTimeout          int    // Connection timeout in seconds
	SSLRootCert             string // Path to root certificate file
	SSLKey                  string // Path to SSL key file
	SSLCert                 string // Path to SSL certificate file

	// SQLite-specific options
	Mode        string // SQLite mode (e.g., "memory", "ro", "rw", "rwc")
	Cache       string // Cache mode (e.g., "shared", "private")
	JournalMode string // Journal mode (e.g., "delete", "truncate", "persist")
	Synchronous string // Synchronization mode (e.g., "off", "normal", "full")

	// Additional custom parameters for flexibility
	CustomParams map[string]string // Any additional parameters
}

// ToDSN generates a driver-specific Data Source Name (DSN) string from the Opts struct.
func (o *Opts) toDSN(driver string) (string, error) {
	switch driver {
	case "mysql":
		params := url.Values{}
		params.Add("charset", o.Charset)
		params.Add("tls", o.TLS)
		params.Add("parseTime", fmt.Sprintf("%t", o.ParseTime))
		params.Add("collation", o.Collation)
		params.Add("allowNativePasswords", fmt.Sprintf("%t", o.AllowNativePasswords))
		params.Add("multiStatements", fmt.Sprintf("%t", o.MultiStatements))
		for k, v := range o.CustomParams {
			params.Add(k, v)
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			o.User, o.Password, o.Host, o.Port, o.Database, params.Encode())
		return dsn, nil

	case "postgres":
		params := url.Values{}
		params.Add("sslmode", o.SSLMode)
		params.Add("search_path", o.SearchPath)
		params.Add("application_name", o.ApplicationName)
		params.Add("fallback_application_name", o.FallbackApplicationName)
		params.Add("connect_timeout", fmt.Sprintf("%d", o.ConnectTimeout))
		params.Add("sslrootcert", o.SSLRootCert)
		params.Add("sslkey", o.SSLKey)
		params.Add("sslcert", o.SSLCert)
		for k, v := range o.CustomParams {
			params.Add(k, v)
		}
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s",
			o.User, o.Password, o.Host, o.Port, o.Database, params.Encode())
		return dsn, nil

	case "sqlite":
		fallthrough
	case "sqlite3":
		params := url.Values{}
		params.Add("mode", o.Mode)
		params.Add("cache", o.Cache)
		params.Add("journal_mode", o.JournalMode)
		params.Add("synchronous", o.Synchronous)
		for k, v := range o.CustomParams {
			params.Add(k, v)
		}
		dsn := fmt.Sprintf("%s?%s", o.Database, params.Encode())
		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}
}

func (o *Opts) setDefaults(driver string) {
	switch driver {
	case "mysql":
		if o.Port == 0 {
			o.Port = 3306
		}
		if o.Charset == "" {
			o.Charset = "utf8mb4"
		}
		if o.TLS == "" {
			o.TLS = "false"
		}
		if o.ParseTime == false {
			o.ParseTime = true
		}
	case "postgres":
		if o.Port == 0 {
			o.Port = 5432
		}
		if o.SSLMode == "" {
			o.SSLMode = "disable"
		}
		if o.ConnectTimeout == 0 {
			o.ConnectTimeout = 10
		}
		if o.ApplicationName == "" {
			o.ApplicationName = "myapp"
		}
	case "sqlite":
		fallthrough
	case "sqlite3":
		if o.Mode == "" {
			o.Mode = "memory"
		}
		if o.Cache == "" {
			o.Cache = "shared"
		}
		if o.Synchronous == "" {
			o.Synchronous = "normal"
		}
	}
	// General defaults for all drivers
	if o.Timeout == 0 {
		o.Timeout = 15
	}
	if o.User == "" {
		o.User = "root"
	}
	if o.Host == "" {
		o.Host = "localhost"
	}
	if o.CustomParams == nil {
		o.CustomParams = make(map[string]string)
	}
}

// Connect creates a connection to the database. The opts string as the second parameter
// is a wrapper of sqlx.Connect but uses a custom Opts struct that it parses into a DSN.
func Connect(driver string, options Opts) Connection {
	options.setDefaults(driver)
	dsn, err := options.toDSN(driver)
	if err != nil {
		panic(err)
	}

	db := sqlx.MustConnect(driver, dsn)
	builder := sq.StatementBuilder.RunWith(db)

	if driver == "postgres" {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}

	return Connection{
		DB:      db,
		Builder: builder,
		driver:  driver,
	}
}
