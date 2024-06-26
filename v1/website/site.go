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
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/bww/go-util/v1/ext"
)

var errCouldNotResolve = errors.New("Could not resolve identity")

var (
	client = &http.Client{}
	log    = slog.With("package", "github.com/bww/go-identify/v1/website")
)

type resolveError struct {
	msg  string
	errs []error
}

func (e resolveError) Error() string {
	return e.msg
}

type Info struct {
	Owner       string // The name of the owner of the site, e.g., a company name, as best we can determine
	Homepage    string // The homepage of the site
	Description string // A brief description of the website and/or the company that manages it
}

// Attempt to infer details about a website from its domain name. This
// convenience interface uses the default resolver.
func IdentifyDomain(cxt context.Context, domain string) (Info, error) {
	return Default().IdentifyDomain(cxt, domain)
}

// Attempt to infer details about a website. This convenience interface uses
// the default resolver.
func IdentifyWebsite(cxt context.Context, link string) (Info, error) {
	return Default().IdentifyWebsite(cxt, link)
}

var (
	defaultResolver     Resolver
	initDefaultResolver sync.Once
)

// Default obtains the shared, default resolver instance. In general, you
// can just use the top level convenience interface if you don't need to
// create your own resolver.
func Default() Resolver {
	initDefaultResolver.Do(func() {
		defaultResolver = New()
	})
	return defaultResolver
}

// Resolver provides methods to resolve details about websites
type Resolver interface {
	IdentifyDomain(cxt context.Context, domain string) (Info, error)
	IdentifyWebsite(cxt context.Context, link string) (Info, error)
}

// New produces a new resolver
func New() Resolver {
	return &standardResolver{}
}

// The standard resolver
type standardResolver struct {
	client *http.Client
}

// Attempt to infer details about a website from its domain name
func (r *standardResolver) IdentifyDomain(cxt context.Context, domain string) (Info, error) {
	var errs []error
	for _, opt := range optionsForDomain(domain) {
		info, err := r.IdentifyWebsite(cxt, fmt.Sprintf("https://%s", opt))
		if err == nil {
			return info, err
		} else {
			errs = append(errs, err)
		}
	}
	return Info{}, resolveError{
		msg:  "Could not resolve identity for domain",
		errs: errs,
	}
}

// Attempt to infer details about a website
func (r *standardResolver) IdentifyWebsite(cxt context.Context, link string) (Info, error) {
	link, err := rootURL(link)
	if err != nil {
		return Info{}, err
	}
	return r.identifyWebsiteWithURL(cxt, link)
}

func (r *standardResolver) identifyWebsiteWithURL(cxt context.Context, link string) (Info, error) {
	req, err := http.NewRequestWithContext(cxt, "GET", link, nil)
	if err != nil {
		return Info{}, nil
	}

	rsp, err := ext.Coalesce(r.client, client, http.DefaultClient).Do(req)
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
		info, err = identifyJSONLD(cxt, info, sel.First().Text()) // ignore errors here; just try alternatives
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
		return info, err
	}
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

// Expand options that are likely to represent the public facing domain given
// the provided input domain
func optionsForDomain(domain string) []string {
	note := map[string]struct{}{domain: struct{}{}}
	opts := []string{domain}

	// add options by removing domain components
	for strings.Count(domain, ".") > 1 {
		if x := strings.Index(domain, "."); x >= 0 {
			domain = domain[x+1:]
			if _, ok := note[domain]; !ok {
				opts = append(opts, domain)
				note[domain] = struct{}{}
			}
		}
	}

	// add options by appending common prefixes
	alt := "www." + domain
	if _, ok := note[alt]; !ok {
		note[alt] = struct{}{}
		opts = append(opts, alt)
	}

	return opts
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
