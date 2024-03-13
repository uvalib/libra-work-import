//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
	"log"
)

func makeEtdObject(namespace string, indir string, excludeFiles bool) (uvaeasystore.EasyStoreObject, error) {

	// import domain metadata
	domainMetadata, err := libraEtdMetadata(indir)
	if err != nil {
		return nil, err
	}

	// import base object
	obj, err := standardObject(namespace, indir)
	if err != nil {
		return nil, err
	}

	// import fields from metadata
	fields, err := libraEtdFields(*domainMetadata)
	if err != nil {
		return nil, err
	}

	// serialize domain metadata
	buf, err := domainMetadata.Payload()
	if err != nil {
		return nil, err
	}

	// create our store metadata object
	metadata := libraMetadata{
		mimeType: domainMetadata.MimeType(),
		payload:  buf,
	}

	// assign fields and serialized metadata
	obj.SetFields(fields)
	obj.SetMetadata(metadata)

	// do we include files?
	if excludeFiles == false {
		// import files if they exist
		blobs, err := importBlobs(namespace, indir)
		if err != nil {
			return nil, err
		}

		if len(blobs) != 0 {
			obj.SetFiles(blobs)
			log.Printf("DEBUG: imported %d files(s) for [%s]", len(blobs), obj.Id())
		} else {
			log.Printf("INFO: no files for [%s]", obj.Id())
		}
	}

	return obj, nil
}

func libraEtdMetadata(indir string) (*librametadata.ETDWork, error) {
	meta := librametadata.ETDWork{}

	buf, err := loadFile(fmt.Sprintf("%s/work.json", indir))
	if err != nil {
		return nil, err
	}

	// convert to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return nil, err
	}

	depositor, err := extractString("depositor", omap["depositor"])
	if err != nil {
		return nil, err
	}

	creator, err := extractString("creator", omap["creator"])
	if err != nil {
		return nil, err
	}

	visibility, err := extractString("embargo_state", omap["embargo_state"])
	if err != nil {
		return nil, err
	}

	embargoRelease, err := extractString("embargo_end_date", omap["embargo_end_date"])
	if err != nil {
		// we can ignore this error
		embargoRelease = ""
	}

	meta.Visibility = visibility

	if len(depositor) != 0 {
		//fields["depositor"] = strings.ReplaceAll(depositor, "@virginia.edu", "")
	}
	if len(creator) != 0 {
		//fields["creator"] = strings.ReplaceAll(creator, "@virginia.edu", "")
	}
	if len(embargoRelease) != 0 && visibility == "restricted" {
		meta.State.EmbargoRelease = embargoRelease
	}

	logEtdMetadata(meta)
	return &meta, nil
}

// extract fields from the domain metadata
func libraEtdFields(meta librametadata.ETDWork) (uvaeasystore.EasyStoreObjectFields, error) {
	fields := uvaeasystore.DefaultEasyStoreFields()

	fields["visibility"] = meta.Visibility

	//if len(meta.Author) != 0 {
	//	fields["depositor"] = strings.ReplaceAll(depositor, "@virginia.edu", "")
	//}
	//if len(creator) != 0 {
	//	fields["creator"] = strings.ReplaceAll(creator, "@virginia.edu", "")
	//}

	if len(meta.State.EmbargoRelease) != 0 && meta.Visibility == "restricted" {
		fields["embargoRelease"] = meta.State.EmbargoRelease
	}

	return fields, nil
}

func logEtdMetadata(meta librametadata.ETDWork) {

	b, _ := meta.Payload()
	fmt.Printf("<%s>\n", string(b))
}

//
// end of file
//
