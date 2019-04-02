package casdm

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

// newRequest é um método acima do http.NewRequest, para reduzir o boilerplate.
func (session Session) newRequest(method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-AccessKey", session.Key)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml; charset=UTF-8")
	req.Header.Set("X-Obj-Attrs", "ref_num, status, summary")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// Shims para realizar a serialização
type ShimStatus struct {
	RelAttr    string `xml:"REL_ATTR,omitempty"`
	CommonName string `xml:"COMMON_NAME,attr,omitempty"`
}
type ShimIn struct {
	XMLName string     `xml:"in"`
	Status  ShimStatus `xml:"status,omitempty"`
	Summary string     `xml:"summary,omitempty"`
}

// convertToIncident converte um in para um Incident
func convertToIncident(i in) (Incident, error) {
	// Converter o ID para um int.
	id, err := strconv.Atoi(i.ID)
	if err != nil {
		return Incident{}, err
	}

	// Converter o Número do incidente para um int.
	number, err := strconv.Atoi(i.RefNum)
	if err != nil {
		return Incident{}, err
	}

	return Incident{
		Status:  i.Status.CommonName,
		Code:    i.Status.RelAttr,
		ID:      id,
		Number:  number,
		Summary: i.Summary,
	}, nil
}

// MarshalRequestBody serializa um Incidente em um <in>, para ser enviado ao CA.
func marshalRequestBody(i Incident) ([]byte, error) {
	x, err := xml.Marshal(ShimIn{
		Status: ShimStatus{
			CommonName: i.Status,
			RelAttr:    i.Code,
		},
		Summary: i.Summary,
	})
	if err != nil {
		return nil, err
	}
	return x, nil
}

// UnmarshalSingle desserializa o XML de um incidente e retorna um Incident
func UnmarshalSingle(resp []byte) (Incident, error) {
	var i in
	if err := xml.Unmarshal(resp, &i); err != nil {
		return Incident{}, err
	}

	return convertToIncident(i)
}

// UnmarshalMultiple desserializa o XML de uma coleção de incidentes e retorna um array de Incidents.
func UnmarshalMultiple(resp []byte) ([]Incident, error) {
	var ci collectionIn

	if err := xml.Unmarshal(resp, &ci); err != nil {
		return nil, err
	}

	// Serializar todos os incidentes para Incidents.
	is := make([]Incident, len(ci.In))
	for n, i := range ci.In {
		ui, err := convertToIncident(i)
		if err != nil {
			is[n] = Incident{}
		}
		is[n] = ui
	}

	return is, nil
}
