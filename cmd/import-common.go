//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"net/http"
	"os"
)

func loadBlobContent(indir string, blob uvaeasystore.EasyStoreBlob) (uvaeasystore.EasyStoreBlob, error) {

	filename := fmt.Sprintf("%s/%s", indir, blob.Name())
	//log.Printf("DEBUG: loading file from %s", filename)
	buf, err := os.ReadFile(filename)
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

//
// end of file
//
