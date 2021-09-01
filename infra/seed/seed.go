package seed

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/isdzulqor/kraicklist/config"
	"github.com/isdzulqor/kraicklist/domain/model"
	"github.com/isdzulqor/kraicklist/external/index"
	errorLib "github.com/isdzulqor/kraicklist/helper/errors"
	"github.com/isdzulqor/kraicklist/helper/logging"
	"github.com/jmoiron/sqlx"
)

func Exec() {
	conf := config.Get()
	conf.PrintPretty()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	logging.InfoContext(ctx, "preparing data seed...")

	ads, err := loadAdsData(conf.Advertisement.MasterDataPath)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	switch conf.IndexerActivated {
	case index.IndexBleve:
		seedDataWithBleve(ctx, ads, conf)
	case index.IndexElastic:
		seedDataWithElastic(ctx, ads, conf)
	case index.IndexSQL:
		seedDataWithSQL(ctx, ads, conf)
	default:
		logging.FatalContext(ctx, "Indexer for %s is invalid", conf.IndexerActivated)
	}

	logging.InfoContext(ctx, "data seed is finished")
}

func loadAdsData(filePath string) (out model.Advertisements, err error) {
	// open file
	file, err := os.Open(filePath)
	if err != nil {
		err = fmt.Errorf("unable to open file due: %v", err)
		return
	}

	defer file.Close()
	// read as gzip
	reader, err := gzip.NewReader(file)
	if err != nil {
		err = fmt.Errorf("unable to initialize gzip reader due: %v", err)
		return
	}

	// read the reader using scanner to contstruct records
	cs := bufio.NewScanner(reader)
	for cs.Scan() {
		ad := model.Advertisement{}
		err = json.Unmarshal(cs.Bytes(), &ad)
		if err != nil {
			continue
		}
		out = append(out, ad)
	}
	return
}

func seedDataWithBleve(ctx context.Context, ads model.Advertisements, conf *config.Config) {
	logging.InfoContext(ctx, "data seeding with bleve index...")

	bleveIndex, err := index.InitBleveIndex(ctx, conf.Advertisement.Bleve.IndexName)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	docs, err := ads.ToBleveDocs()
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
		return
	}

	if docErrors := bleveIndex.BulkIndex(ctx, docs); docErrors != nil {
		logging.ErrContext(ctx, "%v", docErrors.ToError())
	}

	if err := bleveIndex.Close(); err != nil {
		logging.ErrContext(ctx, "%v", err)
	}
}

func seedDataWithElastic(ctx context.Context, ads model.Advertisements, conf *config.Config) {
	logging.InfoContext(ctx, "data seeding with elastic index...")

	esIndex, err := index.InitESIndex(ctx,
		conf.Elastic.Host,
		conf.Elastic.Username,
		conf.Elastic.Password,
		conf.Advertisement.Elastic.IndexName)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	// check es7 cluster readiness with retry
	if err = esIndex.PingWithRetry(conf.Elastic.PingRetry,
		conf.Elastic.PingWaitTime); err != nil {
		logging.WarnContext(ctx, "%v", err)
	}

	if err = esIndex.DeleteIndex(ctx); err != nil {
		logging.WarnContext(ctx, "%v", err)
	}

	docs, err := ads.ToElasticDocs()
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
		return
	}

	docErrors, err := esIndex.BulkIndexDocs(ctx, docs)
	if err != nil {
		logging.ErrContext(ctx, "%v", err)
	}
	if docErrors != nil {
		logging.ErrContext(ctx, "%v", docErrors.ToError())
	}
}

func seedDataWithSQL(ctx context.Context, ads model.Advertisements, conf *config.Config) {
	logging.InfoContext(ctx, "data seeding with SQL DB...")

	sqlDB, err := initPostgreSQL(ctx, conf)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	query, args, err := ads.ToSQLInsert()
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}
	if _, err = sqlDB.DB.ExecContext(ctx, query, args...); err != nil {
		logging.DebugContext(ctx, errorLib.FormatQueryError(query, args...))
		logging.FatalContext(ctx, "%v", err)
	}
	logging.DebugContext(ctx, "successfully inserting %d records", len(ads))
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
