package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/isdzulqor/kraicklist/config"
	"github.com/isdzulqor/kraicklist/domain/handler"
	"github.com/isdzulqor/kraicklist/domain/repository"
	"github.com/isdzulqor/kraicklist/domain/service"
	"github.com/isdzulqor/kraicklist/external/index"
	"github.com/isdzulqor/kraicklist/helper/health"
	"github.com/isdzulqor/kraicklist/helper/logging"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Exec() {
	conf := config.Get()
	conf.PrintPretty()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	handlers := initDependencies(ctx, conf)

	// starting server
	logging.InfoContext(ctx, "Starting HTTP on port %s", conf.Port)
	router := createRouter(ctx, handlers)
	if err := http.ListenAndServe(":"+conf.Port, router); err != nil {
		logging.FatalContext(ctx, "Failed starting HTTP - %v", err)
	}
}

func initDependencies(ctx context.Context, conf *config.Config) handler.Root {
	var (
		err error

		bleveIndex   *index.BleveIndex
		elasticIndex *index.ElasticIndex
		sqlDB        *sqlx.DB

		healthPersistences health.Persistences
	)

	// indexer check
	switch conf.IndexerActivated {
	case index.IndexBleve:
		bleveIndex, err = index.InitBleveIndex(ctx, conf.Advertisement.Bleve.IndexName)
		if err != nil {
			logging.FatalContext(ctx, "%v", err)
		}
	case index.IndexElastic:
		elasticIndex, err = index.InitESIndex(ctx,
			conf.Elastic.Host,
			conf.Elastic.Username,
			conf.Elastic.Password,
			conf.Advertisement.Elastic.IndexName)
		if err != nil {
			logging.FatalContext(ctx, "%v", err)
		}
		// append health persistence
		healthPersistences = append(healthPersistences,
			health.NewPersistence(conf.Advertisement.Elastic.IndexName,
				conf.IndexerActivated, elasticIndex))
	case index.IndexSQL:
		sqlDB, err = initPostgreSQL(ctx, conf)
		if err != nil {
			logging.FatalContext(ctx, "%v", err)
		}
		// append health persistence
		healthPersistences = append(healthPersistences,
			health.NewPersistence(conf.PostgreSQL.Database,
				conf.IndexerActivated, sqlDB))

	default:
		logging.FatalContext(ctx, "Indexer for %s is invalid", conf.IndexerActivated)
	}

	// initialize repo
	adRepo := repository.InitAdvertisement(conf, bleveIndex, elasticIndex, sqlDB)

	// initialize service
	adService := service.InitAdvertisement(adRepo)

	// initialize handlers
	adHandler := handler.InitAdvertisement(conf, adService)
	healthHandler, err := health.NewHealthHandler(&healthPersistences, conf.GracefulShutdownTimeout)
	if err != nil {
		logging.FatalContext(ctx, "failed to init healthHandler")
	}
	healthHandler.WithToken(conf.HealthToken)

	return handler.Root{
		Advertisement: adHandler,
		Health:        healthHandler,
	}
}

func initPostgreSQL(ctx context.Context, conf *config.Config) (db *sqlx.DB, err error) {
	dbConnection := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.PostgreSQL.Host, conf.PostgreSQL.Port, conf.PostgreSQL.Username,
		conf.PostgreSQL.Password, conf.PostgreSQL.Database)

	if db, err = sqlx.Open("postgres", dbConnection); err != nil {
		return
	}

	for i := 30; i > 0; i-- {
		err = db.Ping()
		if err == nil {
			break
		}
		if i == 0 {
			logging.WarnContext(ctx, "Not able to establish connection to database %s", dbConnection)
		}
		logging.WarnContext(ctx, "Could not connect to database. Wait 2 seconds. %d retries left...", i)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return db, err
	}

	db.SetMaxOpenConns(conf.PostgreSQL.ConnectionLimit)
	return db, migrateSQL(db)
}

func migrateSQL(db *sqlx.DB) error {
	migrationPath := "./migration/"
	sqlFiles, err := readSortedFiles(migrationPath)
	if err != nil {
		return err
	}
	for _, v := range sqlFiles {
		sqlString, err := ioutil.ReadFile(migrationPath + v)
		if err != nil {
			return err
		}
		queries := extractQueries(string(sqlString))
		for _, query := range queries {
			if _, err := db.Exec(string(query)); err != nil {
				if strings.Contains(err.Error(), "already exists") {
					continue
				}
				return err
			}
		}
	}
	return nil
}

func readSortedFiles(pathDir string) (fileNames []string, err error) {
	var (
		files []os.FileInfo
		stat  os.FileInfo
	)
	if stat, err = os.Stat(pathDir); err != nil || !stat.IsDir() {
		err = fmt.Errorf("wrong path - %v", err)
		return
	}

	if files, err = ioutil.ReadDir(pathDir); err != nil {
		return
	}
	for _, f := range files {
		if !f.IsDir() {
			fileNames = append(fileNames, f.Name())
		}
	}
	sort.Strings(fileNames)
	return
}

func extractQueries(query string) []string {
	return strings.Split(query, ";")
}
