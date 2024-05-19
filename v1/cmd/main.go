package main

import (
	"github.com/spf13/cobra"
)

var Root = cobra.Command{
	Use: "identify <what>",
}

func main() {
	cobra.CheckErr(Root.Execute())
}
