//
//
//

package main

import (
	//"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"strings"
	"time"
)

type libraOpenSerializer struct {
	namespace string // our namespace
}
type libraEtdSerializer struct {
	namespace string // our namespace
}

//
// Libra Open content deserializer
//

func (impl libraOpenSerializer) BlobDeserialize(i interface{}) (uvaeasystore.EasyStoreBlob, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	// pull the first string from the title array
	title, err := extractFirstString(omap["title"])
	if err != nil {
		return nil, err
	}
	blob := libraBlob{
		name: title,
	}
	return blob, nil
}

func (impl libraOpenSerializer) FieldsDeserialize(i interface{}) (uvaeasystore.EasyStoreObjectFields, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
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

	fields := uvaeasystore.DefaultEasyStoreFields()
	if len(depositor) != 0 {
		fields["depositor"] = strings.ReplaceAll(depositor, "@virginia.edu", "")
	}
	if len(creator) != 0 {
		fields["creator"] = strings.ReplaceAll(creator, "@virginia.edu", "")
	}

	return fields, nil
}

func (impl libraOpenSerializer) MetadataDeserialize(i interface{}) (uvaeasystore.EasyStoreMetadata, error) {

	// all the metadata for now
	metadata := libraMetadata{
		mimeType: "application/json",
		payload:  i.([]byte),
	}
	return metadata, nil
}

func (impl libraOpenSerializer) ObjectDeserialize(i interface{}) (uvaeasystore.EasyStoreObject, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	id, err := extractString(omap["id"])
	if err != nil {
		return nil, err
	}

	o := uvaeasystore.NewEasyStoreObject(impl.namespace, id)
	return o, nil
}

//
// Libra ETD content deserializer
//

func (impl libraEtdSerializer) BlobDeserialize(i interface{}) (uvaeasystore.EasyStoreBlob, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	// pull the first string from the title array
	//fmt.Printf("DEBUG: getting title...\n")
	title, err := extractFirstString(omap["title"])
	if err != nil {
		return nil, err
	}
	blob := libraBlob{
		name: title,
	}
	return blob, nil
}

func (impl libraEtdSerializer) FieldsDeserialize(i interface{}) (uvaeasystore.EasyStoreObjectFields, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("DEBUG: getting depositor...\n")
	depositor, err := extractString(omap["depositor"])
	if err != nil {
		return nil, err
	}

	//fmt.Printf("DEBUG: getting creator...\n")
	creator, err := extractString(omap["creator"])
	if err != nil {
		return nil, err
	}

	//fmt.Printf("DEBUG: getting embargo_state...\n")
	visibility, err := extractString(omap["embargo_state"])
	if err != nil {
		return nil, err
	}

	//fmt.Printf("DEBUG: getting embargo_end_date...\n")
	embargoRelease, err := extractString(omap["embargo_end_date"])
	if err != nil {
		// we can ignore this error
		embargoRelease = ""
	}

	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["visibility"] = visibility

	if len(depositor) != 0 {
		fields["depositor"] = strings.ReplaceAll(depositor, "@virginia.edu", "")
	}
	if len(creator) != 0 {
		fields["creator"] = strings.ReplaceAll(creator, "@virginia.edu", "")
	}
	if len(embargoRelease) != 0 && visibility == "restricted" {
		fields["embargoRelease"] = embargoRelease
	}

	return fields, nil
}

func (impl libraEtdSerializer) MetadataDeserialize(i interface{}) (uvaeasystore.EasyStoreMetadata, error) {

	// all the metadata for now
	metadata := libraMetadata{
		mimeType: "application/json",
		payload:  i.([]byte),
	}
	return metadata, nil
}

func (impl libraEtdSerializer) ObjectDeserialize(i interface{}) (uvaeasystore.EasyStoreObject, error) {

	// convert to a map
	omap, err := interfaceToMap(i)
	if err != nil {
		return nil, err
	}

	id, err := extractString(omap["id"])
	if err != nil {
		return nil, err
	}

	o := uvaeasystore.NewEasyStoreObject(impl.namespace, id)
	return o, nil
}

//
// Custom easystore objects
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
// NOT REQUIRED
//

func (impl libraOpenSerializer) BlobSerialize(b uvaeasystore.EasyStoreBlob) interface{} {
	return nil
}

func (impl libraOpenSerializer) FieldsSerialize(f uvaeasystore.EasyStoreObjectFields) interface{} {
	return nil
}

func (impl libraOpenSerializer) MetadataSerialize(o uvaeasystore.EasyStoreMetadata) interface{} {
	return nil
}

func (impl libraOpenSerializer) ObjectSerialize(o uvaeasystore.EasyStoreObject) interface{} {
	return nil
}

func (impl libraEtdSerializer) BlobSerialize(b uvaeasystore.EasyStoreBlob) interface{} {
	return nil
}

func (impl libraEtdSerializer) FieldsSerialize(f uvaeasystore.EasyStoreObjectFields) interface{} {
	return nil
}

func (impl libraEtdSerializer) MetadataSerialize(o uvaeasystore.EasyStoreMetadata) interface{} {
	return nil
}

func (impl libraEtdSerializer) ObjectSerialize(o uvaeasystore.EasyStoreObject) interface{} {
	return nil
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
