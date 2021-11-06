package main

import (
	"github.com/jvoljvolizka/lovenft/api"
)



func main() {
	app := api.App{
		PatternLocation: "pattern-test",
		MaskLocation:    "lovemasks",
	}
	app.Run(":4242")

}
