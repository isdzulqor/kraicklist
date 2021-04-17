package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"7000"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"` // DEBUG | INFO | WARN | ERROR

	Advertisement struct {
		MasterDataPath string `envconfig:"ADVERTISEMENT_MASTER_DATA_PATH" default:"./data/data.gz"`

		Bleve struct {
			IndexName string `envconfig:"ADVERTISEMENT_BLEVE_INDEX_NAME" default:"kraicklist.bleve"`
		}
	}
}

var once sync.Once
var conf Config

func Get() *Config {
	once.Do(func() {
		if err := envconfig.Process("", &conf); err != nil {
			log.Fatal("Can't load config: ", err)
		}
	})
	return &conf
}

func (c Config) PrintPretty() {
	printTags(reflect.TypeOf(&c).Elem())
	fmt.Println()
}

func printTags(t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Type.Kind() == reflect.Struct {
			printTags(field.Type)
			continue
		}
		column := field.Tag.Get("envconfig")
		value := os.Getenv(column)
		if value == "" {
			value = field.Tag.Get("default")
		}
		fmt.Printf("\033[36m%s: \033[0m%s\n", column, value)
	}
}