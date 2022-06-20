package model

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/tkeel-io/core-broker/pkg/core"
	"github.com/tkeel-io/core-broker/pkg/pagination"
	"github.com/tkeel-io/kit/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	// schema like: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsnFromOSEnvKey = "DSN"
	amqpServer      = "AMQP_SERVER"
)

type WhereOptions func() (query interface{}, args interface{})

var (
	_once      sync.Once
	db         *gorm.DB
	coreClient *core.Client

	AMQPServerAddr = "amqp://localhost:3172"
)

func CoreClient() *core.Client {
	if coreClient != nil {
		return coreClient
	}

	client, err := core.NewCoreClient()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	coreClient = client
	return coreClient
}

func Setup() error {
	var err error

	amqpServerStr := os.Getenv(amqpServer)
	if amqpServerStr != "" {
		AMQPServerAddr = amqpServerStr
	}

	dsn := os.Getenv(dsnFromOSEnvKey)

	// Try to create DB first.
	connectionInfo, dbName := withoutDBConnectionAndDBName(dsn)
	noDBConnection, err := gorm.Open(mysql.Open(connectionInfo), nil)
	if err != nil {
		log.Fatal(err)
	}

	if err = noDBConnection.Exec(createDBSQL(dbName)).Error; err != nil {
		log.Fatal(err)
	}

	// Open the DB
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return db.AutoMigrate(&Subscribe{}, &SubscribeEntities{})
}

func AMQPAddressString(endpoint string) string {
	return AMQPServerAddr + "/" + endpoint
}

func DB() *gorm.DB {
	if db == nil {
		_once.Do(func() {
			if err := Setup(); err != nil {
				log.Error("Setup DB Error: ", err)
				return
			}
		})
	}
	return db
}

func Paginate(find interface{}, page pagination.Page, where interface{}, args ...interface{}) *gorm.DB {
	conditions, fields := page.SearchCondition()

	if conditions != nil && fields == nil {
		return DB().Where(where, args...).Where(conditions).Limit(int(page.Limit())).Offset(int(page.Offset())).Find(find)
	}

	if conditions != nil && fields != nil {
		return DB().Select(fields).Where(where, args...).Where(conditions).Limit(int(page.Limit())).Offset(int(page.Offset())).Find(find)
	}

	if conditions == nil && fields != nil {
		return DB().Select(fields).Where(where, args...).Limit(int(page.Limit())).Offset(int(page.Offset())).Find(find)
	}

	return DB().Where(where, args...).Limit(int(page.Limit())).Offset(int(page.Offset())).Find(find)
}

func ListAll(find interface{}, where interface{}, args ...interface{}) *gorm.DB {
	return DB().Where(where, args...).Find(find)
}

func Count(count *int64, model interface{}, where interface{}, args ...interface{}) *gorm.DB {
	return DB().Model(model).Where(where, args).Count(count)
}

func withoutDBConnectionAndDBName(dsn string) (connection, dbName string) {
	slashIndex := strings.LastIndex(dsn, "/")
	connectionInfo := dsn[:slashIndex+1]
	dbSettings := dsn[slashIndex+1:]
	questionIndex := strings.Index(dbSettings, "?")
	if questionIndex == -1 {
		questionIndex = len(dbSettings)
	}
	dbName = dbSettings[:questionIndex]
	query := dbSettings[questionIndex:]
	connection = connectionInfo + query
	return
}

const createSQLTemplate = "CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;"

func createDBSQL(dbName string) string {
	return fmt.Sprintf(createSQLTemplate, dbName)
}

func CountSubscribeGroupByTentant(model interface{}) map[string]int {
	type res struct {
		Tenant string
		Count  int
	}
	ress := make([]res, 0)
	DB().Model(model).Select("tenant_id as tenant, count(*) as count").Group("tenant").Find(&ress)
	out := make(map[string]int)
	for _, v := range ress {
		out[v.Tenant] = v.Count
	}
	return out
}

func CountSubEntitiesGroupByTentant() map[string]int {
	type res struct {
		Tenant string
		Count  int
	}
	ress := make([]res, 0)
	DB().Raw("SELECT tenant_id as tenant , count(*) as count FROM subscribe_entities join subscribes on subscribe_entities.subscribe_id = subscribes.id  group by tenant_id").Scan(&ress)
	out := make(map[string]int)
	for _, v := range ress {
		out[v.Tenant] = v.Count
	}
	return out
}
