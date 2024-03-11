//
//
//

package main

import (
	"time"
)

//
// Custom easystore metadata objects
//

type libraMetadata struct {
	mimeType string    // mime type (if we know it)
	payload  []byte    // not exposed
	created  time.Time // created time
	modified time.Time // last modified time
}

func (impl libraMetadata) MimeType() string {
	return impl.mimeType
}

func (impl libraMetadata) Payload() ([]byte, error) {
	return impl.payload, nil
}

func (impl libraMetadata) Created() time.Time {
	return impl.created
}

func (impl libraMetadata) Modified() time.Time {
	return impl.modified
}

//
// end of file
//
