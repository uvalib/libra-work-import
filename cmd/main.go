package main

import (
	"flag"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"os"
	"strconv"
)

// global logging level
var logLevel string

// main entry point
func main() {

	var mode string
	var namespace string
	var inDir string
	var importMode string
	var debug bool
	var excludeFiles bool
	var dryRun bool
	var limit int
	var logger *log.Logger

	flag.StringVar(&mode, "mode", "postgres", "Mode, sqlite, postgres, s3")
	flag.StringVar(&namespace, "namespace", "", "Namespace to import")
	flag.StringVar(&inDir, "importdir", "", "Import directory")
	flag.StringVar(&importMode, "importmode", "", "Import mode, either etd or open")
	flag.BoolVar(&debug, "debug", false, "Log debug information")
	flag.BoolVar(&excludeFiles, "nofiles", false, "Do not import files")
	flag.BoolVar(&dryRun, "dryrun", false, "Process but do not actually import")
	flag.IntVar(&limit, "limit", 0, "Number of items to import, 0 for no limit")
	flag.StringVar(&logLevel, "loglevel", "E", "Logging level (D|I|W|E)")
	flag.Parse()

	if debug == true {
		logger = log.Default()
	}

	// validate
	if len(inDir) == 0 {
		logError("must specify import dir")
		os.Exit(1)
	}
	_, err := os.Stat(inDir)
	if err != nil {
		logError(fmt.Sprintf("import dir does not exist or is not readable (%s)", err.Error()))
		os.Exit(1)
	}

	if importMode != "etd" && importMode != "open" {
		logError("import mode must be etd|open")
		os.Exit(1)
	}

	if logLevel != "D" && logLevel != "I" && logLevel != "W" && logLevel != "E" {
		logError("logging level must be D|I|W|E")
		os.Exit(1)
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
		logError(fmt.Sprintf("unsupported mode (%s)", mode))
		os.Exit(1)
	}

	es, err := uvaeasystore.NewEasyStore(config)
	if err != nil {
		logError(fmt.Sprintf("creating easystore (%s)", err.Error()))
		os.Exit(1)
	}

	// important, cleanup properly
	defer es.Close()

	okCount := 0
	errCount := 0
	var obj uvaeasystore.EasyStoreObject

	items, err := os.ReadDir(inDir)
	if err != nil {
		logError(err.Error())
		os.Exit(1)
	}

	if excludeFiles == true {
		logAlways("Excluding file import!!")
	}

	if dryRun == true {
		logAlways("Dryrun, NO import!!")
	}

	// go through our list
	for _, i := range items {
		if i.IsDir() == true {

			// if we are limiting our import count
			if limit != 0 && ((okCount + errCount) >= limit) {
				logDebug(fmt.Sprintf("terminating after %d object(s)", limit))
				break
			}

			dirname := fmt.Sprintf("%s/%s", inDir, i.Name())
			logInfo(fmt.Sprintf("importing from %s", dirname))

			if importMode == "etd" {
				obj, err = makeEtdObject(namespace, dirname, excludeFiles)
			} else {
				obj, err = makeOpenObject(namespace, dirname, excludeFiles)
			}

			if err != nil {
				logError(fmt.Sprintf("creating object (%s), continuing", err.Error()))
				errCount++
				continue
			}

			// if we are configured to import
			if dryRun == false {
				_, err = es.Create(obj)
				if err != nil {
					logError(fmt.Sprintf("importing ns/oid [%s/%s] (%s), continuing", obj.Namespace(), obj.Id(), err.Error()))
					errCount++
					continue
				}
			}

			okCount++
		}
	}

	verb := "imported"
	if dryRun == true {
		verb = "processed"
	}
	logAlways(fmt.Sprintf("terminate normally, %s %d object(s) and %d error(s)", verb, okCount, errCount))
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
