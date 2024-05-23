package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var Root = cobra.Command{
	Use: "identify <what>",
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	cobra.CheckErr(Root.Execute())
}
