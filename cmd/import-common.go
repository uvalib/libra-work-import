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
)

func standardObject(namespace string, indir string) (uvaeasystore.EasyStoreObject, error) {

	buf := loadFile(fmt.Sprintf("%s/work.json", indir))

	// convert to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return nil, err
	}

	id, err := extractString(omap["id"])
	if err != nil {
		return nil, err
	}

	o := uvaeasystore.NewEasyStoreObject(namespace, id)
	return o, nil
}

func makeBlobObject(namespace string, i interface{}) (uvaeasystore.EasyStoreBlob, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	title, err := extractFirstString(omap["title"])
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
	buf := loadFile(filename)

	// we know it's one of these
	b := blob.(libraBlob)

	// set the fields
	b.mimeType = http.DetectContentType(buf)
	b.payload = buf
	return b, nil
}

// terminate on error
func loadFile(filename string) []byte {
	buf, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("ERROR: reading %s (%s)", filename, err.Error())
	}
	return buf
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || errors.Is(err, os.ErrNotExist) == false
}

//
// private methods
//

func interfaceToMap(i interface{}) (map[string]interface{}, error) {

	// assume we are being passed a []byte
	s, ok := i.([]byte)
	if ok != true {
		return nil, fmt.Errorf("%q: %w", "cast error deserializing, interface probably not a []byte", uvaeasystore.ErrDeserialize)
	}

	// deserialize to a map
	var objmap map[string]interface{}
	if err := json.Unmarshal([]byte(s), &objmap); err != nil {
		return nil, fmt.Errorf("%q: %w", err.Error(), uvaeasystore.ErrDeserialize)
	}

	return objmap, nil
}

func extractFirstString(i interface{}) (string, error) {
	fields, ok := i.([]interface{})
	if ok != true {
		return "", fmt.Errorf("%q: %w", "field is not an array", uvaeasystore.ErrDeserialize)
	}
	if len(fields) == 0 {
		return "", nil
	}
	field, ok := fields[0].(string)
	if ok != true {
		return "", fmt.Errorf("%q: %w", "first field is not a string", uvaeasystore.ErrDeserialize)
	}
	return field, nil
}

func extractString(i interface{}) (string, error) {
	field, ok := i.(string)
	if ok != true {
		return "", fmt.Errorf("%q: %w", "field is not a string", uvaeasystore.ErrDeserialize)
	}
	return field, nil
}

//
// end of file
//
