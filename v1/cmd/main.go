package main

import (
	"github.com/spf13/cobra"
)

var cmd = cobra.Command{
	Use: "identify <what>",
}

func main() {
	cobra.CheckErr(cmd.Execute())
}
