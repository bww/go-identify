package website

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

var errCouldNotResolve = errors.New("Could not resolve identity")

var client = &http.Client{}

type Info struct {
	Owner       string // The name of the owner of the site, e.g., a company name, as best we can determine
	Homepage    string // The homepage of the site
	Description string
}

// Attempt to infer details about a website from its domain name
func IdentifyDomain(cxt context.Context, domain string) (Info, error) {
	return IdentifyWebsite(cxt, fmt.Sprintf("https://%s", domain))
}

// Attempt to infer details about a website
func IdentifyWebsite(cxt context.Context, link string) (Info, error) {
	link, err := rootURL(link)
	if err != nil {
		return Info{}, err
	}
	return identifyWebsiteWithURL(cxt, link)
}

// Attempt to infer details about a website
func identifyWebsiteWithURL(cxt context.Context, link string) (Info, error) {
	req, err := http.NewRequestWithContext(cxt, "GET", link, nil)
	if err != nil {
		return Info{}, nil
	}

	rsp, err := client.Do(req)
	if err != nil {
		return Info{}, fmt.Errorf("Could not fech website: %w", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return Info{}, fmt.Errorf("Unexpected response status: %s", rsp.Status)
	}
	if rsp.Body == nil {
		return Info{}, fmt.Errorf("No content returned")
	}
	defer rsp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(rsp.Body)
	if err != nil {
		return Info{}, fmt.Errorf("Could not process document: %w", err)
	}

	return identifyWebsiteWithDocument(cxt, link, doc)
}

func identifyWebsiteWithDocument(cxt context.Context, link string, doc *goquery.Document) (Info, error) {
	var sel *goquery.Selection
	var info Info

	if sel = doc.Find(`head link[rel="canonical"]`); sel.Length() > 0 {
		info.Homepage = sel.First().AttrOr("href", link)
	} else if sel = doc.Find(`head meta[property="og:url"]`); sel.Length() > 0 {
		info.Homepage = sel.First().AttrOr("content", link)
	} else {
		info.Homepage = link
	}

	if sel = doc.Find(`head meta[property="og:site_name"]`); sel.Length() > 0 {
		info.Owner = sel.First().AttrOr("content", "")
	} else if sel = doc.Find(`head meta[name="organization" i]`); sel.Length() > 0 {
		info.Owner = sel.First().AttrOr("content", "")
	}

	if sel = doc.Find(`head meta[name="description" i]`); sel.Length() > 0 {
		info.Description = sel.First().AttrOr("content", "")
	} else if sel = doc.Find(`head meta[property="og:description"]`); sel.Length() > 0 {
		info.Description = sel.First().AttrOr("content", "")
	}

	return info, nil
}

// Attempt to produce a URL representing the root of the input URL
func rootURL(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	u.Path = ""
	return u.String(), nil
}
