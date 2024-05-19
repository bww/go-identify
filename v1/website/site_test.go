package website

import (
	"bytes"
	"context"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestIdentifyWebsite(t *testing.T) {
	tests := []struct {
		Input  string
		Expect Info
		Error  error
	}{
		{
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
