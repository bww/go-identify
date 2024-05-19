package main

import (
	"context"
	"fmt"

	"github.com/bww/go-identify/v1/website"
	"github.com/spf13/cobra"
)

var websiteLink string

func init() {
	ws.Flags().StringVar(&websiteLink, "url", "", "The URL to identify")
	ws.MarkFlagRequired("url")

	Root.AddCommand(ws)
}

var ws = &cobra.Command{
	Use:     "website",
	Aliases: []string{"ws"},
	Short:   "Identify metadata for a website",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := website.IdentifyWebsite(context.Background(), websiteLink)
		cobra.CheckErr(err)
		fmt.Println("   Owner:", info.Owner)
		fmt.Println("Homepage:", info.Homepage)
	},
}
