package ilo_test

import (
	"crypto/tls"
	ilo "github.com/denysvitali/ilo-remote-console/pkg"
	"github.com/go-playground/assert/v2"
	"net/http"
	"os"
	"testing"
)

var tlsSkipHttpClient = http.Client{Transport:
	&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}}

func TestConnect(t *testing.T){
	connection := ilo.NewCustom(&tlsSkipHttpClient)
	err := connection.Connect("10.1.0.3", os.Getenv("ILO_USERNAME"), os.Getenv("ILO_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	info := connection.Info()
	assert.Equal(t, info.ServerName, "sv1")
}

func TestImage(t *testing.T){
	connection := ilo.NewCustom(&tlsSkipHttpClient)
	err := connection.Connect("10.1.0.3", os.Getenv("ILO_USERNAME"), os.Getenv("ILO_PASSWORD"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = connection.GetScreenImage()

	if err != nil {
		t.Fatal(err)
	}
}