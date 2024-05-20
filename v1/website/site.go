package website

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var errCouldNotResolve = errors.New("Could not resolve identity")

var (
	client = &http.Client{}
	log    = slog.With("package", "github.com/bww/go-identify/v1/website")
)

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
	var err error

	if sel = doc.Find(`script[type="application/ld+json"]`); sel.Length() > 0 {
		info, err = identifyJSONLD(cxt, info, sel.Text()) // ignore errors here; just try alternatives
		if err != nil {
			log.Debug(fmt.Sprintf("Could not extract JSON-LD data: %v", err))
		}
	}

	if info.Homepage == "" {
		if sel = doc.Find(`head link[rel="canonical"]`); sel.Length() > 0 {
			info.Homepage = sel.First().AttrOr("href", link)
		} else if sel = doc.Find(`meta[property="og:url"]`); sel.Length() > 0 {
			info.Homepage = sel.First().AttrOr("content", link)
		} else {
			info.Homepage = link
		}
	}

	if info.Owner == "" {
		if sel = doc.Find(`meta[property="og:site_name"]`); sel.Length() > 0 {
			info.Owner = sel.First().AttrOr("content", "")
		} else if sel = doc.Find(`meta[name="organization" i]`); sel.Length() > 0 {
			info.Owner = sel.First().AttrOr("content", "")
		} else if sel = doc.Find(`meta[name="author" i]`); sel.Length() > 0 {
			info.Owner = sel.First().AttrOr("content", "")
		}
	}

	if info.Description == "" {
		if sel = doc.Find(`meta[name="description" i]`); sel.Length() > 0 {
			info.Description = sel.First().AttrOr("content", "")
		} else if sel = doc.Find(`meta[property="og:description"]`); sel.Length() > 0 {
			info.Description = sel.First().AttrOr("content", "")
		}
	}

	return info, nil
}

type jsonLD struct {
	Name        string `json:"name"`
	LegalName   string `json:"legalName"`
	Description string `json:"description"`
}

func identifyJSONLD(cxt context.Context, info Info, data string) (Info, error) {
	var jsonld jsonLD
	err := json.Unmarshal([]byte(strings.TrimSpace(data)), &jsonld)
	if err != nil {
		fmt.Println(">>>>>>>>>>>>>>>>>>>>> YO:", err)
		return info, err
	}
	fmt.Println(">>>>>>>>>>>>>>>>>>>>> YO:", jsonld)
	if jsonld.Name != "" {
		info.Owner = jsonld.Name
	} else if jsonld.LegalName != "" {
		info.Owner = jsonld.LegalName
	}
	if jsonld.Description != "" {
		info.Description = jsonld.Description
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
