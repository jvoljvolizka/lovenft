package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jvoljvolizka/lovenft/api"
	"github.com/jvoljvolizka/lovenft/lambdaapi"
)

func main() {
	islambda := os.Getenv("IS_LAMBDA")
	if islambda == "TRUE" {
		app := lambdaapi.Lapp{
			PatternLocation: os.Getenv("PATTERN_DIR"),
			MaskLocation:    os.Getenv("MASK_DIR"),
			BucketName:      os.Getenv("BUCKET"),
		}
		lambda.Start(app.Generate)
	} else {
		app := api.App{
			PatternLocation: "pattern-test",
			MaskLocation:    "lovemasks",
		}
		app.Run(":4242")
	}

}
