package casdm

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
)

type Status struct {
	Text       string `xml:",chardata"`
	ID         string `xml:"id,attr,omitempty"`
	RelAttr    string `xml:"REL_ATTR,attr,omitempty"`
	CommonName string `xml:"COMMON_NAME,attr,omitempty"`
	Link       Link   `xml:"link,omitempty"`
}

type Link struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr,omitempty"`
	Rel  string `xml:"rel,attr,omitempty"`
}

// Estrutura de um incidente retornado pelo CA.
type in struct {
	Text        string `xml:",chardata"`
	ID          string `xml:"id,attr,omitempty"`
	RelAttr     string `xml:"REL_ATTR,attr,omitempty"`
	CommonName  string `xml:"COMMON_NAME,attr,omitempty"`
	Status      Status `xml:"status,omitempty"`
	Link        Link   `xml:"link,omitempty"`
	RefNum      string `xml:"ref_num,omitempty"`
	Summary     string `xml:"summary,omitempty"`
	Description string `xml:"description,omitempty"`
}

// Estrutura de uma coleção de incidentes retornado pelo CA.
type collectionIn struct {
	XMLName    xml.Name `xml:"collection_in"`
	COUNT      string   `xml:"COUNT,attr"`
	START      string   `xml:"START,attr"`
	TOTALCOUNT string   `xml:"TOTAL_COUNT,attr"`
	Link       struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
	In []in `xml:"in"`
}

// Incident é uma struct que contêm as informações mais relevantes,
// ela deve ser usada com as funções SerializeIncident e SerializeIncidents
type Incident struct {
	Status  string `json:"status,omitempty"`
	Code    string `json:"code,omitempty"`
	Summary string `json:"summary,omitempty"`
	ID      int    `json:"id,omitempty"`
	Number  int    `json:"number,omitempty"`
}

// create Cria um incidente novo no CA com o nome de usuário n, resumo s, e descrição d
func (session Session) create(n string, s string, d string) ([]byte, error) {
	b := bytes.NewBuffer([]byte(""))
	b.Write([]byte(`<in> <customer COMMON_NAME="`))
	err := xml.EscapeText(b, []byte(n))
	if err != nil {
		return nil, err
	}
	b.Write([]byte(`"/> <summary>`))
	err = xml.EscapeText(b, []byte(s))
	if err != nil {
		return nil, err
	}
	b.Write([]byte("</summary> <description>"))
	err = xml.EscapeText(b, []byte(d))
	if err != nil {
		return nil, err
	}
	b.Write([]byte("</description> </in>"))
	if err != nil {
		return nil, err
	}

	fmt.Println(b) // sponge
	return session.newRequest("POST", session.URL+"/in", b)
}

func (session Session) Create(name string, summary string, description string) (Incident, error) {
	b, err := session.create(name, summary, description)
	if err != nil {
		return Incident{}, err
	}

	return UnmarshalSingle(b)
}

// Listar os ultimos n incidentes abertos.
func (session Session) list(n int) ([]byte, error) {
	return session.newRequest("GET",
		session.URL+url.PathEscape("/in"+"?SORT=id DESC&size="+strconv.Itoa(n)),
		bytes.NewBuffer([]byte("")))
}

// Retorna informação sobre um incidente no CA.
func (session Session) get(id string) ([]byte, error) {
	return session.newRequest("GET",
		session.URL+"/in/"+id,
		bytes.NewBuffer([]byte("")))
}

func (session Session) Read(id int) (Incident, error) {
	b, err := session.get(strconv.Itoa(id))
	if err != nil {
		return Incident{}, err
	}

	return UnmarshalSingle(b)
}

func (session Session) put(id int, b []byte) ([]byte, error) {
	return session.newRequest("PUT",
		session.URL+"/in/"+strconv.Itoa(id),
		bytes.NewBuffer(b))
}

func (session Session) Update(id int, in Incident) (Incident, error) {
	req, err := marshalRequestBody(in)
	if err != nil {
		return Incident{}, err
	}
	b, err := session.put(id, req)
	if err != nil {
		return Incident{}, err
	}

	return UnmarshalSingle(b)
}

// List retorna os últimos n incidentes em ordem decrescente.
func (session Session) List(n int) ([]Incident, error) {
	list, err := session.list(n)
	if err != nil {
		return nil, err
	}

	return UnmarshalMultiple(list)
}

// Delete não é possível no CA, então isso só fecha um incidente.
func (session Session) Delete(id int) (Incident, error) {
	return session.Update(id, Incident{
		Status: "Fechado",
	})
}
