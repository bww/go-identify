package main

import (
	"context"
	"fmt"

	"github.com/bww/go-identify/v1/website"
	"github.com/spf13/cobra"
)

var (
	websiteLink   string
	websiteDomain string
)

func init() {
	ws.Flags().StringVar(&websiteLink, "url", "", "The URL to identify")
	ws.Flags().StringVar(&websiteDomain, "domain", "", "The domain to identify")

	Root.AddCommand(ws)
}

var ws = &cobra.Command{
	Use:     "website",
	Aliases: []string{"ws"},
	Short:   "Identify metadata for a website",
	Run: func(cmd *cobra.Command, args []string) {
		var info website.Info
		var err error

		if websiteLink != "" {
			info, err = website.IdentifyWebsite(context.Background(), websiteLink)
		} else if websiteDomain != "" {
			info, err = website.IdentifyDomain(context.Background(), websiteDomain)
		} else {
			err = fmt.Errorf("Specify one of: --url, --domain")
		}
		cobra.CheckErr(err)

		fmt.Println("      Owner:", info.Owner)
		fmt.Println("   Homepage:", info.Homepage)
		fmt.Println("Description:", info.Description)
	},
}
