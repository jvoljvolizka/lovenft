package lambdaapi

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"image/png"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jvoljvolizka/lovenft/imageops"
)

type Lapp struct {
	PatternLocation string
	MaskLocation    string
	BucketName      string
}

func (a *Lapp) Generate(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess)
	path := req.PathParameters["proxy"]

	splittedext := strings.Split(path, ".")

	if len(splittedext) != 2 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("invalid tokenid")
	}
	newNFT, err := imageops.NewImage(splittedext[0])
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, err
	}

	maskBuff := &aws.WriteAtBuffer{}
	_, err = downloader.Download(maskBuff, &s3.GetObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(fmt.Sprintf("%s/%v.png", a.MaskLocation, newNFT.MaskSelector)),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, err
	}
	patternBuff := &aws.WriteAtBuffer{}
	_, err = downloader.Download(patternBuff, &s3.GetObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(fmt.Sprintf("%s/%v.png", a.PatternLocation, newNFT.PatternSelector)),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, err
	}
	patReader := bytes.NewReader(patternBuff.Bytes())
	pattern, err := png.Decode(patReader)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	maskReader := bytes.NewReader(maskBuff.Bytes())
	mask, err := png.Decode(maskReader)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	newNFT.Create(mask, pattern)
	var outBuff bytes.Buffer
	err = png.Encode(&outBuff, newNFT.Image)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Headers:         map[string]string{"Content-type": "image/png"},
		Body:            base64.StdEncoding.EncodeToString(outBuff.Bytes()),
		IsBase64Encoded: true,
	}, nil

}
