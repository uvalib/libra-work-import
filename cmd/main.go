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
	var debug bool
	var excludeFiles bool
	var dryRun bool
	var limit int
	var logger *log.Logger

	flag.StringVar(&mode, "mode", "postgres", "Mode, sqlite, postgres, s3")
	flag.StringVar(&namespace, "namespace", "", "Namespace to import")
	flag.StringVar(&inDir, "importdir", "", "Import directory")
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

	if logLevel != "D" && logLevel != "I" && logLevel != "W" && logLevel != "E" {
		logError("logging level must be D|I|W|E")
		os.Exit(1)
	}

	var implConfig uvaeasystore.EasyStoreImplConfig
	var proxyConfig uvaeasystore.EasyStoreProxyConfig

	// the easystore (or the proxy)
	var es uvaeasystore.EasyStore

	switch mode {
	//	case "sqlite":
	//		implConfig = uvaeasystore.DatastoreSqliteConfig{
	//			DataSource: os.Getenv("SQLITEFILE"),
	//			Log:        logger,
	//		}
	//		es, err = uvaeasystore.NewEasyStore(implConfig)

	case "postgres":
		implConfig = uvaeasystore.DatastorePostgresConfig{
			DbHost:     os.Getenv("DBHOST"),
			DbPort:     asIntWithDefault(os.Getenv("DBPORT"), 0),
			DbName:     os.Getenv("DBNAME"),
			DbUser:     os.Getenv("DBUSER"),
			DbPassword: os.Getenv("DBPASS"),
			DbTimeout:  asIntWithDefault(os.Getenv("DBTIMEOUT"), 0),
			Log:        logger,
		}
		es, err = uvaeasystore.NewEasyStore(implConfig)

	case "s3":
		implConfig = uvaeasystore.DatastoreS3Config{
			Bucket:              os.Getenv("BUCKET"),
			SignerExpireMinutes: asIntWithDefault(os.Getenv("SIGNEXPIRE"), 60),
			SignerAccessKey:     os.Getenv("SIGNER_ACCESS_KEY"),
			SignerSecretKey:     os.Getenv("SIGNER_SECRET_KEY"),
			DbHost:              os.Getenv("DBHOST"),
			DbPort:              asIntWithDefault(os.Getenv("DBPORT"), 0),
			DbName:              os.Getenv("DBNAME"),
			DbUser:              os.Getenv("DBUSER"),
			DbPassword:          os.Getenv("DBPASS"),
			DbTimeout:           asIntWithDefault(os.Getenv("DBTIMEOUT"), 0),
			BusName:             os.Getenv("BUSNAME"),
			SourceName:          os.Getenv("SOURCENAME"),
			Log:                 logger,
		}
		es, err = uvaeasystore.NewEasyStore(implConfig)

	case "proxy":
		proxyConfig = uvaeasystore.ProxyConfigImpl{
			ServiceEndpoint: os.Getenv("ESENDPOINT"),
			Log:             logger,
		}
		es, err = uvaeasystore.NewEasyStoreProxy(proxyConfig)

	default:
		logError(fmt.Sprintf("unsupported mode (%s)", mode))
		os.Exit(1)
	}

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
	total := len(items)
	for ix, i := range items {
		if i.IsDir() == true {

			// if we are limiting our import count
			if limit != 0 && ((okCount + errCount) >= limit) {
				logDebug(fmt.Sprintf("terminating after %d object(s)", limit))
				break
			}

			dirname := fmt.Sprintf("%s/%s", inDir, i.Name())
			logInfo(fmt.Sprintf("importing from %s (%d of %d)", dirname, ix+1, total))

			obj, err = makeEtdObject(namespace, dirname, excludeFiles)

			if err != nil {
				logError(fmt.Sprintf("creating object (%s), continuing", err.Error()))
				errCount++
				continue
			}

			// if we are configured to import
			if dryRun == false {
				_, err = es.ObjectCreate(obj)
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
