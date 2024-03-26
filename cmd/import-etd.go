//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
	"log"
	"sort"
	"strconv"
	"strings"
)

func makeEtdObject(namespace string, indir string, excludeFiles bool) (uvaeasystore.EasyStoreObject, error) {

	// import domain metadata plus any extras that we need that dont have a place in the metadata
	domainMetadata, domainExtras, err := libraEtdMetadata(indir)
	if err != nil {
		return nil, err
	}

	// import base object
	obj, err := standardObject(namespace, indir)
	if err != nil {
		return nil, err
	}

	// import fields from metadata
	fields, err := libraEtdFields(domainMetadata, domainExtras)
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

func libraEtdMetadata(indir string) (librametadata.ETDWork, importExtras, error) {
	meta := librametadata.ETDWork{}
	extra := importExtras{}

	buf, err := loadFile(fmt.Sprintf("%s/work.json", indir))
	if err != nil {
		return meta, extra, err
	}

	// convert to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return meta, extra, err
	}

	meta.Degree, err = extractString("degree", omap["degree"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Visibility, err = extractString("embargo_state", omap["embargo_state"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Title, err = extractFirstString("title", omap["title"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	// meta.Author handled below

	// meta.Advisors handled below

	meta.Abstract, err = extractString("description", omap["description"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.License, err = extractFirstString("rights", omap["rights"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Keywords, err = extractStringArray("keyword", omap["keyword"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Language, err = extractString("language", omap["language"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.RelatedURLs, err = extractStringArray("related_url", omap["related_url"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.PublicationDate, err = extractString("date_published", omap["date_published"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Sponsors, err = extractStringArray("sponsoring_agency", omap["sponsoring_agency"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Notes, err = extractString("notes", omap["notes"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.depositor, err = extractString("depositor", omap["depositor"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Author, err = libraEtdAuthor(omap)
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Advisors, err = libraEtdAdvisors(omap)
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	meta.Visibility, err = extractString("embargo_state", omap["embargo_state"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.createDate, err = extractString("date_created", omap["date_created"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	//
	// extra stuff that does not form part of the metadata but is stored in the object fields
	//

	extra.adminNotes, err = extractStringArray("admin_notes", omap["admin_notes"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.doi, err = extractString("permanent_url", omap["permanent_url"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.depositor, err = extractString("depositor", omap["depositor"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.embargoRelease, err = extractString("embargo_end_date", omap["embargo_end_date"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.createDate, err = extractString("date_created", omap["date_created"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	extra.source, err = extractString("work_source", omap["work_source"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		//return nil, err
	}

	//logEtdMetadata(meta)
	return meta, extra, nil
}

// extract fields from the domain metadata plus the extras
func libraEtdFields(meta librametadata.ETDWork, extra importExtras) (uvaeasystore.EasyStoreObjectFields, error) {
	fields := uvaeasystore.DefaultEasyStoreFields()

	// all imported items get these
	fields["disposition"] = "imported"
	fields["draft"] = "false"
	fields["invitation-sent"] = "imported"
	fields["submitted-sent"] = "imported"

	// all imported ETD's get these
	fields["sis-sent"] = "imported"

	if len(extra.adminNotes) != 0 {
		fields["admin-notes"] = strings.Join(extra.adminNotes, " ")
	}

	if len(meta.Author.ComputeID) != 0 {
		fields["author"] = meta.Author.ComputeID
	}

	if len(extra.createDate) != 0 {
		fields["create-date"] = extra.createDate
	}

	if len(extra.depositor) != 0 {
		fields["depositor"] = extra.depositor
	}

	if len(extra.doi) != 0 {
		// cleanup the DOI
		doi := strings.Replace(extra.doi, "https://doi.org/", "", 1)
		doi = strings.Replace(doi, "http://dx.doi.org/", "", 1)
		fields["doi"] = fmt.Sprintf("https://doi.org/%s", extra.doi)
	}

	if len(extra.embargoRelease) != 0 && meta.Visibility == "restricted" {
		fields["embargo-release"] = extra.embargoRelease
	}

	if len(extra.source) != 0 {
		fields["source-id"] = extra.source
		fields["source"] = strings.Trim(
			strings.Split(extra.source, ":")[0], " ")
	}

	if len(meta.Visibility) != 0 {
		fields["visibility"] = meta.Visibility
	}

	return fields, nil
}

func libraEtdAuthor(omap map[string]interface{}) (librametadata.StudentData, error) {

	author := librametadata.StudentData{}
	var err error

	author.ComputeID, err = extractString("author_email", omap["author_email"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
	}
	author.ComputeID = strings.Replace(author.ComputeID, "@virginia.edu", "", -1)

	author.FirstName, err = extractString("author_first_name", omap["author_first_name"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
	}

	author.LastName, err = extractString("author_last_name", omap["author_last_name"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
	}

	// FIXME ???
	author.Program, err = extractString("degree", omap["degree"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
	}

	author.Institution, err = extractString("author_institution", omap["author_institution"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
	}

	return author, nil
}

func libraEtdAdvisors(omap map[string]interface{}) ([]librametadata.ContributorData, error) {

	advisors := make([]librametadata.ContributorData, 0)
	contributors, err := extractStringArray("contributor", omap["contributor"])
	if err != nil {
		log.Printf("WARNING: %s", err.Error())
		return advisors, nil
	}

	local := make([]LocalContributorData, 0)
	for _, str := range contributors {

		// cleanup and split
		str = strings.Replace(str, "\r\n", "\n", -1)
		sarray := strings.Split(str, "\n")
		if len(sarray) == 6 {
			var advisor LocalContributorData
			advisor.Index, err = strconv.Atoi(sarray[0])
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
				continue
			}
			advisor.ComputeID = sarray[1]
			advisor.FirstName = sarray[2]
			advisor.LastName = sarray[3]
			advisor.Department = sarray[4]
			advisor.Institution = sarray[5]

			// and add to the list
			local = append(local, advisor)
		} else {
			log.Printf("WARNING: badly formatted contributor entry")
		}
	}

	// we need to ensure that these are included in the order they were added
	sort.Sort(ContributorSorter(local))
	for _, p := range local {
		advisors = append(advisors, librametadata.ContributorData{
			ComputeID:   strings.TrimSpace(p.ComputeID), // for some reason
			FirstName:   p.FirstName,
			LastName:    p.LastName,
			Department:  p.Department,
			Institution: p.Institution,
		})
	}
	return advisors, nil
}

func logEtdMetadata(meta librametadata.ETDWork) {

	b, _ := meta.Payload()
	fmt.Printf("<%s>\n", string(b))
}

//
// end of file
//
