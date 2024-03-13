package main

import (
	"flag"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"os"
	"strconv"
)

// main entry point
func main() {

	var mode string
	var namespace string
	var inDir string
	var importMode string
	var debug bool
	var excludeFiles bool
	var limit int
	var logger *log.Logger

	flag.StringVar(&mode, "mode", "postgres", "Mode, sqlite, postgres, s3")
	flag.StringVar(&namespace, "namespace", "", "Namespace to import")
	flag.StringVar(&inDir, "importdir", "", "Import directory")
	flag.StringVar(&importMode, "importmode", "", "Import mode, either etd or open")
	flag.BoolVar(&debug, "debug", false, "Log debug information")
	flag.BoolVar(&excludeFiles, "nofiles", false, "Do not import files")
	flag.IntVar(&limit, "limit", 0, "Number of items to import, 0 for no limit")
	flag.Parse()

	if debug == true {
		logger = log.Default()
	}

	// validate
	if len(inDir) == 0 {
		log.Fatalf("ERROR: must specify import dir")
	}
	_, err := os.Stat(inDir)
	if err != nil {
		log.Fatalf("ERROR: import dir does not exist or is not readable (%s)", err.Error())
	}

	if importMode != "etd" && importMode != "open" {
		log.Fatalf("ERROR: import mode must be 'etd' or 'open'")
	}

	var config uvaeasystore.EasyStoreConfig

	switch mode {
	case "sqlite":
		config = uvaeasystore.DatastoreSqliteConfig{
			DataSource: os.Getenv("SQLITEFILE"),
			Log:        logger,
		}
	case "postgres":
		config = uvaeasystore.DatastorePostgresConfig{
			DbHost:     os.Getenv("DBHOST"),
			DbPort:     asIntWithDefault(os.Getenv("DBPORT"), 0),
			DbName:     os.Getenv("DBNAME"),
			DbUser:     os.Getenv("DBUSER"),
			DbPassword: os.Getenv("DBPASS"),
			DbTimeout:  asIntWithDefault(os.Getenv("DBTIMEOUT"), 0),
			Log:        logger,
		}
	case "s3":
		config = uvaeasystore.DatastoreS3Config{
			Bucket:     os.Getenv("BUCKET"),
			DbHost:     os.Getenv("DBHOST"),
			DbPort:     asIntWithDefault(os.Getenv("DBPORT"), 0),
			DbName:     os.Getenv("DBNAME"),
			DbUser:     os.Getenv("DBUSER"),
			DbPassword: os.Getenv("DBPASS"),
			DbTimeout:  asIntWithDefault(os.Getenv("DBTIMEOUT"), 0),
			Log:        logger,
		}
	default:
		log.Fatalf("ERROR: unsupported mode (%s)", mode)
	}

	es, err := uvaeasystore.NewEasyStore(config)
	if err != nil {
		log.Fatalf("ERROR: creating easystore (%s)", err.Error())
	}

	okCount := 0
	errCount := 0
	var obj uvaeasystore.EasyStoreObject

	items, err := os.ReadDir(inDir)
	if err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}

	if excludeFiles == true {
		log.Printf("INFO: EXCLUDING file import!!")
	}

	// go through our list
	for _, i := range items {
		if i.IsDir() == true {

			// if we are limiting our import count
			if limit != 0 && ((okCount + errCount) >= limit) {
				log.Printf("DEBUG: terminating after %d object(s)", limit)
				break
			}

			dirname := fmt.Sprintf("%s/%s", inDir, i.Name())
			log.Printf("DEBUG: importing from %s", dirname)

			if importMode == "etd" {
				obj, err = makeEtdObject(namespace, dirname, excludeFiles)
			} else {
				obj, err = makeOpenObject(namespace, dirname, excludeFiles)
			}

			if err != nil {
				log.Printf("ERROR: creating object (%s), continuing", err.Error())
				errCount++
				continue
			}

			_, err = es.Create(obj)
			if err != nil {
				log.Printf("ERROR: importing ns/oid [%s/%s] (%s), continuing", obj.Namespace(), obj.Id(), err.Error())
				errCount++
			} else {
				okCount++
			}
		}
	}

	log.Printf("INFO: terminate normally, imported %d object(s) and %d error(s)", okCount, errCount)
}

func asIntWithDefault(str string, def int) int {
	if len(str) == 0 {
		return def
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return def
	}
	return i
}

//
// end of file
//
