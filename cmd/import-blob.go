//
//
//

package main

import (
	"time"
)

//
// Custom easystore blob implementation
//

type libraBlob struct {
	name     string    // source file name
	mimeType string    // mime type (if we know it)
	payload  []byte    // not exposed
	created  time.Time // created time
	modified time.Time // last modified time
}

func (impl libraBlob) Name() string {
	return impl.name
}

func (impl libraBlob) MimeType() string {
	return impl.mimeType
}

func (impl libraBlob) Url() string {
	return "https://does.not.work.fu"
}

func (impl libraBlob) Payload() ([]byte, error) {
	return impl.payload, nil
}

func (impl libraBlob) Created() time.Time {
	return impl.created
}

func (impl libraBlob) Modified() time.Time {
	return impl.modified
}

//
// end of file
//
