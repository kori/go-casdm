package casdm

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type Session struct {
	URL            string `json:"url"`
	Key            string `json:"key"`
	ExpirationDate string `json:"expiration_date"`
}

// NewSession retorna uma Session com uma chave de acesso para o CASDM e a URL de qual CASDM está sendo usado.
func NewSession(url string, credentials string) (Session, error) {
	cred := base64.StdEncoding.EncodeToString([]byte(credentials))
	req, err := http.NewRequest("POST", url+"/rest_access", bytes.NewBuffer([]byte("<rest_access/>")))
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Authorization", "Basic "+cred)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Session{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Session{}, err
	}

	var ra struct {
		Link           string `xml:"link"`
		AccessKey      string `xml:"access_key"`
		ExpirationDate string `xml:"expiration_date"`
	}
	xml.Unmarshal(body, &ra)

	return Session{
		URL:            url,
		Key:            ra.AccessKey,
		ExpirationDate: ra.ExpirationDate,
	}, nil
}
