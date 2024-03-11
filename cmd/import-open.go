//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"strings"
)

func makeOpenObject(namespace string, indir string) (uvaeasystore.EasyStoreObject, error) {

	// import base object
	obj, err := standardObject(namespace, indir)
	if err != nil {
		return nil, err
	}

	// import fields
	fields, err := libraOpenFields(namespace, indir)
	if err != nil {
		return nil, err
	}

	obj.SetFields(fields)

	// import metadata
	metadata, err := libraOpenMetadata(namespace, indir)
	if err != nil {
		return nil, err
	}
	obj.SetMetadata(metadata)

	// import files if they exist
	blobs := make([]uvaeasystore.EasyStoreBlob, 0)
	ix := 0
	var blob uvaeasystore.EasyStoreBlob
	exists := fileExists(fmt.Sprintf("%s/fileset-1.json", indir))
	for exists == true {

		// load the blob content
		buf := loadFile(fmt.Sprintf("%s/fileset-%d.json", indir, ix+1))
		blob, err = makeBlobObject(namespace, buf)
		if err != nil {
			return nil, err
		}
		blob, err = loadBlobContent(indir, blob)
		if err != nil {
			return nil, err
		}

		// and add to the list
		blobs = append(blobs, blob)
		ix++
		exists = fileExists(fmt.Sprintf("%s/fileset-%d.json", indir, ix+1))
	}

	if len(blobs) != 0 {
		obj.SetFiles(blobs)
		log.Printf("INFO: ==> imported %d blob(s) for [%s]", ix, obj.Id())
	} else {
		log.Printf("INFO: no files for [%s]", obj.Id())
	}

	return obj, nil
}

func libraOpenMetadata(namespace string, indir string) (uvaeasystore.EasyStoreMetadata, error) {
	// TBD
	return libraMetadata{}, nil
}

func libraOpenFields(namespace string, indir string) (uvaeasystore.EasyStoreObjectFields, error) {

	fields := uvaeasystore.DefaultEasyStoreFields()

	buf := loadFile(fmt.Sprintf("%s/work.json", indir))

	// convert to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return nil, err
	}

	depositor, err := extractString(omap["depositor"])
	if err != nil {
		return nil, err
	}

	creator, err := extractFirstString(omap["creator"])
	if err != nil {
		return nil, err
	}

	if len(depositor) != 0 {
		fields["depositor"] = strings.ReplaceAll(depositor, "@virginia.edu", "")
	}
	if len(creator) != 0 {
		fields["creator"] = strings.ReplaceAll(creator, "@virginia.edu", "")
	}

	// get extras that do not appear in the work.json file
	visibility := libraOpenVisibility(namespace, indir)
	embargoRelease := libraOpenEmbargo(namespace, indir)
	if len(visibility) != 0 {
		fields["visibility"] = visibility
	}
	if len(embargoRelease) != 0 && visibility == "restricted" {
		fields["embargoRelease"] = embargoRelease
	}

	return fields, nil
}

func libraOpenVisibility(namespace string, indir string) string {

	fname := fmt.Sprintf("%s/visibility.json", indir)
	exists := fileExists(fname)
	if exists == false {
		// assume no visibility information
		return ""
	}

	buf := loadFile(fname)
	omap, err := interfaceToMap(buf)
	if err != nil {
		// assume no visibility information
		return ""
	}
	str, err := extractString(omap["visibility"])
	if err != nil {
		// assume no visibility information
		return ""
	}
	return str
}

func libraOpenEmbargo(namespace string, indir string) string {

	fname := fmt.Sprintf("%s/embargo.json", indir)
	exists := fileExists(fname)
	if exists == false {
		// assume no embargo information
		return ""
	}
	buf := loadFile(fname)
	omap, err := interfaceToMap(buf)
	if err != nil {
		// assume no embargo information
		return ""
	}
	str, err := extractString(omap["embargo_release_date"])
	if err != nil {
		// assume no embargo information
		return ""
	}
	return str
}

//
// end of file
//
