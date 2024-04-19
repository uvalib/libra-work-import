//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
	"sort"
	"strconv"
	"strings"
)

var etdRights = map[string]string{

	"Attribution 4.0 International (CC BY)":                        "http://creativecommons.org/licenses/by/4.0/",
	"All rights reserved (no additional license for public reuse)": "",
}

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
			logDebug(fmt.Sprintf("imported %d files(s) for [%s]", len(blobs), obj.Id()))
		} else {
			logInfo(fmt.Sprintf("no files for [%s]", obj.Id()))
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

	meta.Department, err = extractString("department", omap["department"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Degree, err = extractString("degree", omap["degree"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Title, err = extractFirstString("title", omap["title"])
	if err != nil {
		logWarning(err.Error())
	}

	// meta.Author handled below

	// meta.Advisors handled below

	meta.Abstract, err = extractString("description", omap["description"])
	if err != nil {
		logWarning(err.Error())
	}

	rights, err := extractFirstString("rights", omap["rights"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.License, meta.LicenseURL = libraEtdRights(rights)

	meta.Keywords, err = extractStringArray("keyword", omap["keyword"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Language, err = extractString("language", omap["language"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.RelatedURLs, err = extractStringArray("related_url", omap["related_url"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.pubDate, err = extractString("date_published", omap["date_published"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Sponsors, err = extractStringArray("sponsoring_agency", omap["sponsoring_agency"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Notes, err = extractString("notes", omap["notes"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.depositor, err = extractString("depositor", omap["depositor"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Author, err = libraEtdAuthor(omap)
	if err != nil {
		logWarning(err.Error())
	}

	meta.Advisors, err = libraEtdAdvisors(omap)
	if err != nil {
		logWarning(err.Error())
	}

	extra.defaultVis, err = extractString("embargo_state", omap["embargo_state"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.createDate, err = extractString("date_created", omap["date_created"])
	if err != nil {
		logWarning(err.Error())
	}

	//
	// extra stuff that does not form part of the metadata but is stored in the object fields
	//

	extra.adminNotes, err = extractStringArray("admin_notes", omap["admin_notes"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.doi, err = extractString("permanent_url", omap["permanent_url"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.depositor, err = extractString("depositor", omap["depositor"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.embargoRelease, err = extractString("embargo_end_date", omap["embargo_end_date"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.embargoVisDuring = extra.defaultVis
	extra.embargoVisAfter = "open"

	extra.createDate, err = extractString("date_created", omap["date_created"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.source, err = extractString("work_source", omap["work_source"])
	if err != nil {
		logWarning(err.Error())
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
		fields["depositor"] = strings.Replace(extra.depositor, "@virginia.edu", "", -1)
	}

	// we may adjust this later if we have embargo information
	if len(extra.defaultVis) != 0 {
		fields["default-visibility"] = extra.defaultVis
	}

	if len(extra.doi) != 0 {
		// cleanup the DOI
		doi := strings.Replace(extra.doi, "https://doi.org/", "", 1)
		doi = strings.Replace(doi, "http://dx.doi.org/", "", 1)
		fields["doi"] = fmt.Sprintf("https://doi.org/%s", doi)
	}

	// embargo visibility calculations
	if expectedEmbargoDateFormat(extra.embargoRelease) {
		fields["embargo-release"] = extra.embargoRelease
		if inTheFuture(extra.embargoRelease) == true {
			if len(extra.embargoVisDuring) != 0 {
				fields["default-visibility"] = extra.embargoVisDuring
			}
		}

		if len(extra.embargoVisAfter) != 0 {
			fields["embargo-release-visibility"] = extra.embargoVisAfter
		} else {
			fields["embargo-release-visibility"] = extra.defaultVis
		}
	}

	// field name change
	if fields["default-visibility"] == "authenticated" {
		fields["default-visibility"] = "uva"
	}
	if fields["embargo-release-visibility"] == "authenticated" {
		fields["embargo-release-visibility"] = "uva"
	}

	if len(extra.pubDate) != 0 {
		fields["publish-date"] = extra.pubDate
	}

	if len(extra.source) != 0 {
		fields["source-id"] = extra.source
		fields["source"] = strings.Trim(
			strings.Split(extra.source, ":")[0], " ")
	}

	return fields, nil
}

func libraEtdAuthor(omap map[string]interface{}) (librametadata.ContributorData, error) {

	author := librametadata.ContributorData{}
	var err error

	author.ComputeID, err = extractString("author_email", omap["author_email"])
	if err != nil {
		logWarning(err.Error())
	}
	author.ComputeID = strings.Replace(author.ComputeID, "@virginia.edu", "", -1)

	author.FirstName, err = extractString("author_first_name", omap["author_first_name"])
	if err != nil {
		logWarning(err.Error())
	}

	author.LastName, err = extractString("author_last_name", omap["author_last_name"])
	if err != nil {
		logWarning(err.Error())
	}

	// FIXME ???
	author.Program, err = extractString("degree", omap["degree"])
	if err != nil {
		logWarning(err.Error())
	}

	author.Institution, err = extractString("author_institution", omap["author_institution"])
	if err != nil {
		logWarning(err.Error())
	}

	return author, nil
}

func libraEtdAdvisors(omap map[string]interface{}) ([]librametadata.ContributorData, error) {

	advisors := make([]librametadata.ContributorData, 0)
	contributors, err := extractStringArray("contributor", omap["contributor"])
	if err != nil {
		logWarning(err.Error())
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
				logWarning(err.Error())
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
			logWarning("badly formatted contributor entry")
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

func libraEtdRights(rights string) (string, string) {

	url, ok := etdRights[rights]
	if ok == true {
		return rights, url
	}
	return rights, ""
}

func logEtdMetadata(meta librametadata.ETDWork) {

	b, _ := meta.Payload()
	fmt.Printf("<%s>\n", string(b))
}

//
// end of file
//
