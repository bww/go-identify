package website

import (
	"bytes"
	"context"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestDomainOptions(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Expect []string
	}{
		{
			Name:  "TLD (this isn't really valid, but...)",
			Input: "google",
			Expect: []string{
				"google",
				"www.google",
			},
		},
		{
			Name:  "Root domain",
			Input: "google.com",
			Expect: []string{
				"google.com",
				"www.google.com",
			},
		},
		{
			Name:  "Email domain",
			Input: "email.google.com",
			Expect: []string{
				"email.google.com",
				"google.com",
				"www.google.com",
			},
		},
		{
			Name:  "Many sub-domains",
			Input: "x1.y2.email.google.com",
			Expect: []string{
				"x1.y2.email.google.com",
				"y2.email.google.com",
				"email.google.com",
				"google.com",
				"www.google.com",
			},
		},
	}
	for _, e := range tests {
		assert.Equal(t, e.Expect, optionsForDomain(e.Input))
	}
}

func TestIdentifyWebsite(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Expect Info
		Error  error
	}{
		{
			Name: "First choice options",
			Input: `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
		<meta name='robots' content='index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' />
		<!-- This site is optimized with the Yoast SEO plugin v22.7 - https://yoast.com/wordpress/plugins/seo/ -->
		<title>Home - EverHealth</title>
		<meta name="descriPTION" content="No tiers, no restrictions, all the benefits." />
		<meta property="og:locale" content="en_US" />
		<meta property="og:type" content="website" />
		<meta property="og:title" content="Home - EverHealth" />
		<meta property="og:description" content="Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations." />
		<meta property="og:url" content="https://www.everhealth.com/meta" />
		<meta property="og:site_name" content="EverHealth" />
		<meta property="article:modified_time" content="2024-05-14T19:13:39+00:00" />
		<meta property="og:image" content="https://www.everhealth.com/wp-content/uploads/everhealth-logo.svg" />
		<meta name="twitter:card" content="summary_large_image" />
	</head>
</html>`,
			Expect: Info{
				Owner:       "EverHealth",
				Homepage:    "https://www.everhealth.com/meta",
				Description: "No tiers, no restrictions, all the benefits.",
			},
		},
		{
			Name: "Fallback options",
			Input: `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
		<meta name='robots' content='index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' />
		<!-- This site is optimized with the Yoast SEO plugin v22.7 - https://yoast.com/wordpress/plugins/seo/ -->
		<title>Home - EverHealth</title>
		<link rel="canonical" href="https://www.everhealth.com/link" />
		<meta property="og:locale" content="en_US" />
		<meta property="og:type" content="website" />
		<meta property="og:title" content="Home - EverHealth" />
		<meta property="og:description" content="Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations." />
		<meta property="og:url" content="https://www.everhealth.com/meta" />
    <meta content="University of California, San Diego" name="ORGANIzation"/>
		<meta property="article:modified_time" content="2024-05-14T19:13:39+00:00" />
		<meta property="og:image" content="https://www.everhealth.com/wp-content/uploads/everhealth-logo.svg" />
		<meta name="twitter:card" content="summary_large_image" />
	</head>
</html>`,
			Expect: Info{
				Owner:       "University of California, San Diego",
				Homepage:    "https://www.everhealth.com/link",
				Description: "Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations.",
			},
		},
		{
			Name: "Duplciates only use the first encountered instance",
			Input: `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
		<meta name='robots' content='index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' />
		<!-- This site is optimized with the Yoast SEO plugin v22.7 - https://yoast.com/wordpress/plugins/seo/ -->
		<title>Home - EverHealth</title>
		<link rel="canonical" href="https://www.everhealth.com/link" />
		<meta property="og:locale" content="en_US" />
		<meta property="og:type" content="website" />
		<meta property="og:title" content="Home - EverHealth" />
		<meta property="og:description" content="Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations." />
		<meta property="og:description" content="We're in Oregon now." />
		<meta property="og:url" content="https://www.everhealth.com/meta" />
    <meta content="University of California, San Diego" name="ORGANIzation"/>
    <meta content="University of Oregon, Portland" name="ORGANIzation"/>
		<meta property="article:modified_time" content="2024-05-14T19:13:39+00:00" />
		<meta property="og:image" content="https://www.everhealth.com/wp-content/uploads/everhealth-logo.svg" />
		<meta name="twitter:card" content="summary_large_image" />
	</head>
</html>`,
			Expect: Info{
				Owner:       "University of California, San Diego",
				Homepage:    "https://www.everhealth.com/link",
				Description: "Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations.",
			},
		},
		{
			Name: "Malformed meta tags not in head",
			Input: `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
		<meta name='robots' content='index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' />
		<!-- This site is optimized with the Yoast SEO plugin v22.7 - https://yoast.com/wordpress/plugins/seo/ -->
		<title>Home - EverHealth</title>
		<link rel="canonical" href="https://www.everhealth.com/link" />
	</head>
	<body></body>
	<meta property="og:locale" content="en_US" />
	<meta property="og:type" content="website" />
	<meta property="og:title" content="Home - EverHealth" />
	<meta property="og:description" content="Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations." />
	<meta property="og:url" content="https://www.everhealth.com/meta" />
	<meta content="University of California, San Diego" name="ORGANIzation"/>
	<meta property="article:modified_time" content="2024-05-14T19:13:39+00:00" />
	<meta property="og:image" content="https://www.everhealth.com/wp-content/uploads/everhealth-logo.svg" />
	<meta name="twitter:card" content="summary_large_image" />
</html>`,
			Expect: Info{
				Owner:       "University of California, San Diego",
				Homepage:    "https://www.everhealth.com/link",
				Description: "Reimagining the Way You Work Our simplified, user-centric software can streamline daily operations.",
			},
		},
		{
			Name: "JSON-LD source",
			Input: `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
		<meta name='robots' content='index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' />
		<!-- This site is optimized with the Yoast SEO plugin v22.7 - https://yoast.com/wordpress/plugins/seo/ -->
		<title>Home - EverHealth</title>
		<link rel="canonical" href="https://www.everhealth.com/link" />
    <script type="application/ld+json">
        {
          "@context": "https://schema.org/",
          "@type": "Corporation",
          "@id": "#Corporation",
          "url": "https://www.lumen.me/",
          "legalName": "Lumen",
          "name": "Lumen Inc",
          "description": "Lumen is the world’s first hand-held, portable device to accurately measure metabolism. Once available only to top athletes, in hospitals and clinics, metabolic testing is now available to everyone.",
          "image": "https://www.lumen.me/assets/Pages/home/science-device.png",
          "logo": "https://www.lumen.me/assets/og1.png",
          "email": "support@lumen.me",
          "address": {
            "@type": "PostalAddress",
            "streetAddress": "325 Hudson St., 4th Floor",
            "addressLocality": "Manhattan",
            "addressRegion": "New York",
            "addressCountry": "United States",
            "postalCode": "10013"
          },
          "sameAs": [
            "https://www.lumen.me/about",
            "https://www.facebook.com/Lumen.me/",
            "https://www.youtube.com/channel/UC3XkEyGUMXfRhZcB0Ve_fQQ",
            "https://www.instagram.com/lumen.me/",
            "https://www.linkedin.com/company/lumen-me/",
            "https://www.pinterest.com/MyLumen",
            "https://apps.apple.com/us/app/lumen-metabolism-tracker/id1395149502",
            "https://play.google.com/store/apps/details?id=com.metaflow.lumen"
          ]
        }
    </script>
	</head>
</html>`,
			Expect: Info{
				Owner:       "Lumen Inc",
				Homepage:    "https://www.everhealth.com/link",
				Description: "Lumen is the world’s first hand-held, portable device to accurately measure metabolism. Once available only to top athletes, in hospitals and clinics, metabolic testing is now available to everyone.",
			},
		},
	}
	for _, e := range tests {
		doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(e.Input)))
		if !assert.NoError(t, err) {
			continue
		}
		info, err := identifyWebsiteWithDocument(context.TODO(), "https://default.com", doc)
		if e.Error != nil {
			assert.ErrorIs(t, err, e.Error)
		} else if assert.NoError(t, err) {
			assert.Equal(t, e.Expect, info)
		}
	}
}
