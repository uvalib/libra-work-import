//
//
//

package main

import (
	"encoding/json"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/libra-metadata"
	"log"
	"sort"
	"strings"
)

func makeOpenObject(namespace string, indir string, excludeFiles bool) (uvaeasystore.EasyStoreObject, error) {

	// import domain metadata
	domainMetadata, err := libraOpenMetadata(indir)
	if err != nil {
		return nil, err
	}

	// import base object
	obj, err := standardObject(namespace, indir)
	if err != nil {
		return nil, err
	}

	// extract fields from metadata
	fields, err := libraOpenFields(*domainMetadata)
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

func libraOpenMetadata(indir string) (*librametadata.OAWork, error) {
	var err error
	meta := librametadata.OAWork{}

	buf, err := loadFile(fmt.Sprintf("%s/work.json", indir))
	if err != nil {
		return nil, err
	}

	// convert buffer to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return nil, err
	}

	// meta.Visibility handled below

	meta.ResourceType, err = extractString("resource_type", omap["resource_type"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Title, err = extractFirstString("title", omap["title"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	// meta.Authors handled below

	meta.Abstract, err = extractString("abstract", omap["abstract"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	// TODO: meta.License

	meta.Languages, err = extractStringArray("language", omap["language"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Keywords, err = extractStringArray("keyword", omap["keyword"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	// meta.Contributors handled below

	meta.Publisher, err = extractString("publisher", omap["publisher"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Citation, err = extractString("source_citation", omap["source_citation"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.PublicationDate, err = extractString("published_date", omap["published_date"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Sponsors, err = extractStringArray("sponsoring_agency", omap["sponsoring_agency"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
	}

	meta.RelatedURLs, err = extractStringArray("related_url", omap["related_url"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Notes, err = extractString("notes", omap["notes"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	//
	// get extras that do not appear in the work.json file
	//

	meta.Authors, err = libraOpenAuthors(indir)
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Contributors, err = libraOpenContributors(indir)
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Visibility = libraOpenVisibility(indir)
	embargoRelease := libraOpenEmbargo(indir)
	if len(embargoRelease) != 0 && meta.Visibility == "restricted" {
		meta.State.EmbargoRelease = embargoRelease
	}

	//logOpenMetadata(meta)
	return &meta, nil
}

// extract fields from the domain metadata
func libraOpenFields(meta librametadata.OAWork) (uvaeasystore.EasyStoreObjectFields, error) {

	fields := uvaeasystore.DefaultEasyStoreFields()

	// FIXME
	if len(meta.Authors) != 0 && len(meta.Authors[0].ComputeID) != 0 {
		fields["depositor"] = meta.Authors[0].ComputeID
	}
	if len(meta.Authors) != 0 && len(meta.Authors[0].ComputeID) != 0 {
		fields["author"] = meta.Authors[0].ComputeID
	}

	if len(meta.Visibility) != 0 {
		fields["visibility"] = meta.Visibility
	}
	if len(meta.State.EmbargoRelease) != 0 && meta.Visibility == "restricted" {
		fields["embargoRelease"] = meta.State.EmbargoRelease
	}

	return fields, nil
}

func libraOpenAuthors(indir string) ([]librametadata.ContributorData, error) {
	return loadContributorData(indir, "%s/author-%d.json")
}

func libraOpenContributors(indir string) ([]librametadata.ContributorData, error) {
	return loadContributorData(indir, "%s/contributor-%d.json")
}

func loadContributorData(indir string, template string) ([]librametadata.ContributorData, error) {

	local := make([]LocalContributorData, 0)
	ix := 1
	exists := fileExists(fmt.Sprintf(template, indir, ix))
	for exists == true {

		// load the file contents
		buf, err := loadFile(fmt.Sprintf(template, indir, ix))
		if err != nil {
			return nil, err
		}

		// decode into local structure
		var person LocalContributorData
		err = json.Unmarshal(buf, &person)
		if err != nil {
			return nil, err
		}

		// and add to the list
		local = append(local, person)
		ix++
		exists = fileExists(fmt.Sprintf(template, indir, ix))
	}

	// we need to ensure that these are included in the order they were added
	people := make([]librametadata.ContributorData, 0)
	sort.Sort(ContributorSorter(local))
	for _, p := range local {
		people = append(people, librametadata.ContributorData{
			ComputeID:   strings.TrimSpace(p.ComputeID), // for some reason
			FirstName:   p.FirstName,
			LastName:    p.LastName,
			Department:  p.Department,
			Institution: p.Institution,
		})
	}
	return people, nil
}

func libraOpenVisibility(indir string) string {

	fname := fmt.Sprintf("%s/visibility.json", indir)
	exists := fileExists(fname)
	if exists == false {
		// assume no visibility information
		return ""
	}

	buf, err := loadFile(fname)
	if err != nil {
		// assume no visibility information
		return ""
	}
	omap, err := interfaceToMap(buf)
	if err != nil {
		// assume no visibility information
		return ""
	}
	str, err := extractString("visibility", omap["visibility"])
	if err != nil {
		// assume no visibility information
		return ""
	}
	return str
}

func libraOpenEmbargo(indir string) string {

	fname := fmt.Sprintf("%s/embargo.json", indir)
	exists := fileExists(fname)
	if exists == false {
		// assume no embargo information
		return ""
	}
	buf, err := loadFile(fname)
	if err != nil {
		// assume no embargo information
		return ""
	}
	omap, err := interfaceToMap(buf)
	if err != nil {
		// assume no embargo information
		return ""
	}
	str, err := extractString("embargo_release_date", omap["embargo_release_date"])
	if err != nil {
		// assume no embargo information
		return ""
	}
	return str
}

func logOpenMetadata(meta librametadata.OAWork) {

	b, _ := meta.Payload()
	fmt.Printf("<%s>\n", string(b))
}

//
// end of file
//
