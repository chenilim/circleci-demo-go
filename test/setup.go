package test

import (
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/CircleCI-Public/circleci-demo-go/service"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
	"github.com/stretchr/testify/require"
)

// Env provides access to all services used in tests, like the database, our server, and an HTTP client for performing
// HTTP requests against the test server.
type Env struct {
	T          *testing.T
	DB         *service.Database
	Server     *service.Server
	HttpServer *httptest.Server
	Client     service.Client
}

// Close must be called after each test to ensure the Env is properly destroyed.
func (env *Env) Close() {
	env.HttpServer.Close()
	env.DB.Close()
}

// SetupEnv creates a new test environment, including a clean database and an instance of our HTTP service.
func SetupEnv(t *testing.T) *Env {
	db := SetupDB(t)
	server := service.NewServer(db)
	httpServer := httptest.NewServer(server)
	return &Env{
		T:          t,
		DB:         db,
		Server:     server,
		HttpServer: httpServer,
		Client:     service.NewClient(httpServer.URL),
	}
}

// SetupDB initializes a test database, performing all migrations.
func SetupDB(t *testing.T) *service.Database {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("CWD: %v\n", path)

	path, err = os.Executable()
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("exec path: %v\n", path)

	home := os.Getenv("HOME")
	fmt.Printf("HOME: %v\n", home)

	workspace := os.Getenv("GITHUB_WORKSPACE")
	fmt.Printf("GITHUB_WORKSPACE: %v\n", workspace)

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	fmt.Printf("GITHUB_EVENT_PATH: %v\n", eventPath)

	databaseUrl := os.Getenv("CONTACTS_DB_URL")
	require.NotEmpty(t, databaseUrl, "CONTACTS_DB_URL must be set!")

	sqlFiles := "./db/migrations"
	if sqlFilesEnv := os.Getenv("CONTACTS_DB_MIGRATIONS"); sqlFilesEnv != "" {
		sqlFiles = sqlFilesEnv
	}
	allErrors, ok := migrate.ResetSync(databaseUrl, sqlFiles)
	require.True(t, ok, "Failed to migrate database %v", allErrors)

	db, err := sql.Open("postgres", databaseUrl)
	require.NoError(t, err, "Error opening database")

	return &service.Database{db}
}
