//
//
//

package main

import (
	"errors"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"os"
)

func makeObjectFromOpen(serializer uvaeasystore.EasyStoreSerializer, indir string) (uvaeasystore.EasyStoreObject, error) {

	buf, err := os.ReadFile(fmt.Sprintf("%s/work.json", indir))
	if err != nil {
		log.Fatalf("ERROR: reading file (%s)", err.Error())
	}

	// import base object
	obj, err := serializer.ObjectDeserialize(buf)
	if err != nil {
		return nil, err
	}

	// import fields
	fields, err := serializer.FieldsDeserialize(buf)
	if err != nil {
		return nil, err
	}

	// get extras that do not appear in the work.json file
	visibility := libraOpenVisibility(indir)
	embargoRelease := libraOpenEmbargo(indir)
	if len(visibility) != 0 {
		fields["visibility"] = visibility
	}
	if len(embargoRelease) != 0 && visibility == "restricted" {
		fields["embargoRelease"] = embargoRelease
	}
	obj.SetFields(fields)

	// import metadata
	metadata, err := serializer.MetadataDeserialize(buf)
	if err != nil {
		return nil, err
	}
	obj.SetMetadata(metadata)

	// import files if they exist
	buf, err = os.ReadFile(fmt.Sprintf("%s/fileset-1.json", indir))
	if err == nil {

		// for each possible blob file
		blobs := make([]uvaeasystore.EasyStoreBlob, 0)
		ix := 0
		var blob uvaeasystore.EasyStoreBlob
		buf, err = os.ReadFile(fmt.Sprintf("%s/fileset-%d.json", indir, ix+1))
		for err == nil {

			blob, err = serializer.BlobDeserialize(buf)
			if err != nil {
				return nil, err
			}

			blob, err = loadBlobContent(indir, blob)
			if err != nil {
				return nil, err
			}

			blobs = append(blobs, blob)
			ix++
			buf, err = os.ReadFile(fmt.Sprintf("%s/fileset-%d.json", indir, ix+1))
		}
		if errors.Is(err, os.ErrNotExist) == true {
			obj.SetFiles(blobs)
			log.Printf("INFO: ==> imported %d blob(s) for [%s]", ix, obj.Id())
		} else {
			return nil, err
		}
	} else {
		if errors.Is(err, os.ErrNotExist) == true {
			log.Printf("INFO: no files for [%s]", obj.Id())
		} else {
			return nil, err
		}
	}

	return obj, nil
}

func libraOpenVisibility(indir string) string {
	buf, err := os.ReadFile(fmt.Sprintf("%s/visibility.json", indir))
	if err != nil {
		// assume no visibility information
		return ""
	}
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

func libraOpenEmbargo(indir string) string {
	buf, err := os.ReadFile(fmt.Sprintf("%s/embargo.json", indir))
	if err != nil {
		// assume no embargo information
		return ""
	}
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
