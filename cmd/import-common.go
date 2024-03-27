//
//
//

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"net/http"
	"os"
	"time"
)

var location, _ = time.LoadLocation("America/New_York")

// the structure for importing is slightly different
type LocalContributorData struct {
	Index       int    `json:"index"`
	ComputeID   string `json:"computing_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Department  string `json:"department"`
	Institution string `json:"institution"`
}

// libra ETD and some libra Open have this
type EmbargoDetails struct {
	VisibilityDuring string `json:"visibility_during_embargo"`
	VisibilityAfter  string `json:"visibility_after_embargo"`
	ReleaseDate      string `json:"embargo_release_date"`
}

type importExtras struct {
	adminNotes       []string
	createDate       string
	defaultVis       string // default visibility
	depositor        string
	doi              string
	embargoRelease   string // embargo release date (if appropriate)
	embargoVisDuring string // visibility during embargo
	embargoVisAfter  string // visibility after embargo
	source           string
}

type ContributorSorter []LocalContributorData

func (c ContributorSorter) Len() int           { return len(c) }
func (c ContributorSorter) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ContributorSorter) Less(i, j int) bool { return c[i].Index < c[j].Index }

func standardObject(namespace string, indir string) (uvaeasystore.EasyStoreObject, error) {

	buf, err := loadFile(fmt.Sprintf("%s/work.json", indir))
	if err != nil {
		return nil, err
	}

	// convert to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return nil, err
	}

	id, err := extractString("id", omap["id"])
	if err != nil {
		return nil, err
	}

	o := uvaeasystore.NewEasyStoreObject(namespace, id)
	return o, nil
}

func importBlobs(namespace string, indir string) ([]uvaeasystore.EasyStoreBlob, error) {
	blobs := make([]uvaeasystore.EasyStoreBlob, 0)
	ix := 1
	var blob uvaeasystore.EasyStoreBlob
	exists := fileExists(fmt.Sprintf("%s/fileset-%d.json", indir, ix))
	for exists == true {

		// load the blob content
		buf, err := loadFile(fmt.Sprintf("%s/fileset-%d.json", indir, ix))
		if err != nil {
			return nil, err
		}

		blob, err = makeBlobObject(namespace, buf)
		if err != nil {
			return nil, err
		}
		// some cases where we have bad files
		if len(blob.Name()) != 0 {
			blob, err = loadBlobContent(indir, blob)
			if err != nil {
				return nil, err
			}

			// and add to the list
			blobs = append(blobs, blob)
		} else {
			log.Printf("WARNING: empty blob name, skipping")
		}
		ix++
		exists = fileExists(fmt.Sprintf("%s/fileset-%d.json", indir, ix))
	}
	return blobs, nil
}

func makeBlobObject(namespace string, i interface{}) (uvaeasystore.EasyStoreBlob, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	title, err := extractFirstString("title", omap["title"])
	if err != nil {
		return nil, err
	}
	blob := libraBlob{
		name: title,
	}

	return blob, nil
}

func loadBlobContent(indir string, blob uvaeasystore.EasyStoreBlob) (uvaeasystore.EasyStoreBlob, error) {

	filename := fmt.Sprintf("%s/%s", indir, blob.Name())
	buf, err := loadFile(filename)
	if err != nil {
		return nil, err
	}

	// we know it's one of these
	b := blob.(libraBlob)

	// set the fields
	b.mimeType = http.DetectContentType(buf)
	b.payload = buf
	return b, nil
}

func loadFile(filename string) ([]byte, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || errors.Is(err, os.ErrNotExist) == false
}

func loadEmbargo(indir string) (*EmbargoDetails, error) {
	fname := fmt.Sprintf("%s/embargo.json", indir)
	exists := fileExists(fname)
	if exists == false {
		return nil, os.ErrNotExist
	}

	buf, err := loadFile(fname)
	if err != nil {
		return nil, err
	}

	var embargo EmbargoDetails
	err = json.Unmarshal(buf, &embargo)
	if err != nil {
		return nil, err
	}

	return &embargo, nil
}

func interfaceToMap(i interface{}) (map[string]interface{}, error) {

	// assume we are being passed a []byte
	s, ok := i.([]byte)
	if ok != true {
		return nil, fmt.Errorf("%q: %w", "cast error deserializing, interface probably not a []byte", uvaeasystore.ErrDeserialize)
	}

	// deserialize to a map
	var objmap map[string]interface{}
	if err := json.Unmarshal(s, &objmap); err != nil {
		return nil, fmt.Errorf("%q: %w", err.Error(), uvaeasystore.ErrDeserialize)
	}

	return objmap, nil
}

func extractFirstString(name string, i interface{}) (string, error) {
	fields, err := extractStringArray(name, i)
	if err != nil {
		return "", err
	}
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], nil
}

func extractStringArray(name string, i interface{}) ([]string, error) {
	result := make([]string, 0)
	fields, ok := i.([]interface{})
	if ok != true {
		return result, fmt.Errorf("%q: %w", fmt.Sprintf("%s is not an array", name), uvaeasystore.ErrDeserialize)
	}
	for _, f := range fields {
		field, ok := f.(string)
		if ok != true {
			return result, fmt.Errorf("%q: %w", fmt.Sprintf("%s array element is not a string", name), uvaeasystore.ErrDeserialize)
		}
		result = append(result, field)
	}
	return result, nil
}

func extractString(name string, i interface{}) (string, error) {
	field, ok := i.(string)
	if ok != true {
		return "", fmt.Errorf("%q: %w", fmt.Sprintf("%s is not a string", name), uvaeasystore.ErrDeserialize)
	}
	return field, nil
}

func inTheFuture(datetime string) bool {
	if len(datetime) == 0 {
		return false
	}

	format := "2006-01-02T15:04:05+00:00" // yeah, crap right
	dt, err := time.ParseInLocation(format, datetime, location)
	if err != nil {
		log.Printf("ERROR: bad date format [%s] (%s)", datetime, err.Error())
		return false
	}

	return dt.After(time.Now())
}

func expectedEmbargoDateFormat(datetime string) bool {
	if len(datetime) == 0 {
		return false
	}

	format := "2006-01-02T15:04:05+00:00" // yeah, crap right
	_, err := time.ParseInLocation(format, datetime, location)
	if err != nil {
		log.Printf("ERROR: bad date format [%s] (%s)", datetime, err.Error())
		return false
	}

	return true
}

//
// end of file
//
