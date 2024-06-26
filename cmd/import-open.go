//
//
//

package main

import (
	"encoding/json"
	"fmt"
	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/libra-metadata"
	"sort"
	"strconv"
	"strings"
)

type Right struct {
	url  string
	term string
}

var openRights = []Right{
	{url: "http://creativecommons.org/licenses/by/4.0/",
		term: "Attribution 4.0 International (CC BY)"},

	{url: "http://creativecommons.org/licenses/by-nd/4.0",
		term: "Attribution-NoDerivatives 4.0 International (CC BY-ND)"},

	{url: "http://creativecommons.org/licenses/by-sa/4.0/",
		term: "Attribution-ShareAlike 4.0 International (CC BY-SA)"},

	{url: "http://creativecommons.org/licenses/by-nc/4.0/",
		term: "Attribution-NonCommercial 4.0 International (CC BY-NC)"},

	{url: "http://creativecommons.org/licenses/by-nc-nd/4.0/",
		term: "Attribution-NonCommercial-NoDerivatives 4.0 International (CC BY-NC-ND)"},

	{url: "http://creativecommons.org/licenses/by-nc-sa/4.0",
		term: "Attribution-NonCommercial-ShareAlike 4.0 International (CC BY-NC-SA)"},

	{url: "http://creativecommons.org/publicdomain/zero/1.0/",
		term: "CC0 1.0 Universal"},

	{url: "",
		term: "All rights reserved (no additional license for public reuse)"},

	{url: "https://creativecommons.org/licenses/by/2.0/",
		term: "Attribution 2.0 Generic (CC BY)"},
}

func makeOpenObject(namespace string, indir string, excludeFiles bool) (uvaeasystore.EasyStoreObject, error) {

	// import domain metadata plus any extras that we need that dont have a place in the metadata
	domainMetadata, domainExtras, err := libraOpenMetadata(indir)
	if err != nil {
		return nil, err
	}

	// import base object
	obj, err := standardObject(namespace, indir)
	if err != nil {
		return nil, err
	}

	// extract fields from metadata and extra stuff
	fields, err := libraOpenFields(domainMetadata, domainExtras)
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

func libraOpenMetadata(indir string) (librametadata.OAWork, importExtras, error) {
	var err error
	meta := librametadata.OAWork{}
	extra := importExtras{}

	buf, err := loadFile(fmt.Sprintf("%s/work.json", indir))
	if err != nil {
		return meta, extra, err
	}

	// convert buffer to a map
	omap, err := interfaceToMap(buf)
	if err != nil {
		return meta, extra, err
	}

	// meta.Visibility handled below

	meta.ResourceType, err = extractString("resource_type", omap["resource_type"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Title, err = extractFirstString("title", omap["title"])
	if err != nil {
		logWarning(err.Error())
	}

	// meta.Authors handled below

	meta.Abstract, err = extractString("abstract", omap["abstract"])
	if err != nil {
		logWarning(err.Error())
	}

	// meta.License handled below

	meta.Languages, err = extractStringArray("language", omap["language"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Keywords, err = extractStringArray("keyword", omap["keyword"])
	if err != nil {
		logWarning(err.Error())
	}

	// meta.Contributors handled below

	meta.Publisher, err = extractString("publisher", omap["publisher"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Citation, err = extractString("source_citation", omap["source_citation"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.PublicationDate, err = extractString("published_date", omap["published_date"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Sponsors, err = extractStringArray("sponsoring_agency", omap["sponsoring_agency"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.RelatedURLs, err = extractStringArray("related_url", omap["related_url"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.Notes, err = extractString("notes", omap["notes"])
	if err != nil {
		logWarning(err.Error())
	}

	//
	// other stuff that does not appear in the work.json file
	//

	meta.Authors, err = libraOpenAuthors(indir)
	if err != nil {
		logWarning(err.Error())
	}

	meta.Contributors, err = libraOpenContributors(indir)
	if err != nil {
		logWarning(err.Error())
	}

	rightsIndex, err := extractFirstString("rights", omap["rights"])
	if err != nil {
		logWarning(err.Error())
	}

	meta.License, meta.LicenseURL = libraOpenRights(rightsIndex)

	//
	// extra stuff that does not form part of the metadata but is stored in the object fields
	//

	extra.adminNotes, err = extractStringArray("admin_notes", omap["admin_notes"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.depositor, err = extractString("depositor", omap["depositor"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.doi, err = extractString("doi", omap["doi"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.defaultVis = libraOpenVisibility(indir)
	embargo, err := loadEmbargo(indir)
	if err == nil {
		extra.embargoVisAfter = embargo.VisibilityAfter
		extra.embargoVisDuring = embargo.VisibilityDuring
		extra.embargoRelease = embargo.ReleaseDate
	}

	extra.createDate, err = extractString("date_created", omap["date_created"])
	if err != nil {
		logWarning(err.Error())
	}

	extra.source, err = extractString("work_source", omap["work_source"])
	if err != nil {
		logWarning(err.Error())
	}

	//logOpenMetadata(meta)
	return meta, extra, nil
}

// extract fields from the domain metadata plus the extras
func libraOpenFields(meta librametadata.OAWork, extra importExtras) (uvaeasystore.EasyStoreObjectFields, error) {

	fields := uvaeasystore.DefaultEasyStoreFields()

	// all imported items get these
	fields["disposition"] = "imported"
	fields["draft"] = "false"
	fields["invitation-sent"] = "imported"
	fields["submitted-sent"] = "imported"

	if len(extra.adminNotes) != 0 {
		fields["admin-notes"] = strings.Join(extra.adminNotes, " ")
	}

	if len(meta.Authors) != 0 && len(meta.Authors[0].ComputeID) != 0 {
		fields["author"] = meta.Authors[0].ComputeID
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
		// turn the DOI into a URL
		extra.doi = strings.Replace(extra.doi, "doi:", "", 1)
		fields["doi"] = fmt.Sprintf("https://doi.org/%s", extra.doi)
	}

	// embargo visibility calculations
	if len(extra.embargoRelease) != 0 {
		extra.embargoRelease = cleanupDate(extra.embargoRelease)
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

	if len(meta.PublicationDate) != 0 {
		date := cleanupDate(meta.PublicationDate)
		if len(date) != 0 {
			fields["publish-date"] = date
		}
	}

	if len(meta.ResourceType) != 0 {
		fields["resource-type"] = meta.ResourceType
	}

	if len(extra.source) != 0 {
		fields["source-id"] = extra.source
		fields["source"] = strings.Trim(
			strings.Split(extra.source, ":")[0], " ")
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

func libraOpenRights(rightsIndex string) (string, string) {

	ix, err := strconv.Atoi(rightsIndex)
	if err != nil {
		logError(err.Error())
		// assume no rights information
		return "", ""
	}

	// incorrect number
	if ix >= len(openRights) {
		logError(fmt.Sprintf("%s too big", rightsIndex))
		// assume no rights information
		return "", ""
	}

	return openRights[ix].term, openRights[ix].url
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
