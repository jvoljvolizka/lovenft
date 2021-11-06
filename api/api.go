package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jvoljvolizka/lovenft/imageops"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func apiErrorHandle500(w http.ResponseWriter, err error) {

	var jerror JsonError
	jerror.Error = err.Error()
	respondWithJSON(w, http.StatusInternalServerError, jerror)

}

func apiErrorHandle400(w http.ResponseWriter, err error) {

	var jerror JsonError
	jerror.Error = err.Error()
	respondWithJSON(w, http.StatusBadRequest, jerror)

}

type App struct {
	Router          *mux.Router
	PatternLocation string
	MaskLocation    string
}

type JsonError struct {
	Error string
}

func (a *App) Run(addr string) {
	a.Initialize()

	srv := &http.Server{
		Handler: a.Router,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("cool cool")
	fmt.Println(addr)

	srv.ListenAndServe()
}

func (a *App) Initialize() {

	a.Router = mux.NewRouter()

	a.Router.HandleFunc("/{tokenid}", a.createNFTHandler).Methods("GET")

}

func (a *App) createNFTHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	newNFT, err := imageops.NewImage(params["tokenid"])
	fmt.Println(newNFT)
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	lovefiles, err := ioutil.ReadDir(a.MaskLocation)
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	if newNFT.MaskSelector > uint8(len(lovefiles)-1) {

		err = fmt.Errorf("invalid tokenID")
		apiErrorHandle400(w, err)
		return
	}
	patternfiles, err := ioutil.ReadDir(a.PatternLocation)
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	if newNFT.PatternSelector > uint8(len(patternfiles)-1) {
		err = fmt.Errorf("invalid tokenID")
		apiErrorHandle400(w, err)
		return
	}
	f, err := os.Open(a.MaskLocation + "/" + lovefiles[newNFT.MaskSelector].Name())
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	mask, err := png.Decode(f)
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	f, err = os.Open(a.PatternLocation + "/" + patternfiles[newNFT.PatternSelector].Name())
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	pattern, err := png.Decode(f)
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	newNFT.Create(mask, pattern)
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, newNFT.Image)
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	_, err = w.Write(buffer.Bytes())
	if err != nil {
		apiErrorHandle500(w, err)
		return
	}
	return
}
