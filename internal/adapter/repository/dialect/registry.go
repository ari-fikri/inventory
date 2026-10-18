package dialect

import (
	"fmt"
	"strings"
)

var defaultRegistry = map[Driver]Dialect{
	DriverSQLite:   NewSQLiteDialect(),
	DriverPostgres: NewPostgresDialect(),
	DriverOracle:   NewOracleDialect(),
}

// Get returns the dialect registered for the specified driver name.
func Get(driver Driver) (Dialect, error) {
	d, ok := defaultRegistry[Driver(strings.ToLower(string(driver)))]
	if !ok {
		return nil, fmt.Errorf("unsupported database driver '%s'", driver)
	}
	return d, nil
}
