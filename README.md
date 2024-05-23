# `go-identify` identifies things on the web
This package provides routines that extract information about things on the web using whichever _first-party_ metadata is available. Generally this means fetching a webpage and inspecting website headers.

Currently, `go-identify` can probe metadata relating to:

* Websites, and
* Domains (which it uses to guess where it might find a corresponding website).

In future, it is imagined that `go-identify` may also support:

* Email addresses,
* Twitter, Facebook, and other social media handles,
* Stuff like that.

## Using the API
```go
import "github.com/bww/go-identify/v1/website"

info, err := website.IdentifyDomain(context.TODO(), "treno.io")
if err != nil {
  // ...
}

fmt.Printf("The domain treno.io is owned by %s\n", info.Owner)
// The domain treno.io is owned by Treno
```

## Using the CLI
`go-identify` is primarily a Go package, but a CLI wrapper is provided to do the same thing on the command line:

```sh
$ go install ./v1/cmd/metaid
$ metaid website --domain treno.io
      Owner: Treno
   Homepage: https://www.treno.io/
Description: Treno is an observability platform that provides monitoring, metrics, and visualizations that allow you to observe, analyze, and improve software delivery.
```

