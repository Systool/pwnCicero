package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/encoding/charmap"
)

type PwnNetworking struct {
	uUID   string
	devID  string
	client http.Client
}

func NewNet(deviceID string, proxy *url.URL) (*PwnNetworking, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	encDeviceID, err := charmap.ISO8859_1.NewEncoder().Bytes([]byte(deviceID))
	if err != nil {
		return nil, err
	}

	hasher := sha1.New()
	if _, err := hasher.Write(encDeviceID); err != nil {
		return nil, err
	}

	net := &PwnNetworking{
		id.String(),
		hex.EncodeToString(hasher.Sum(nil)),
		http.Client{},
	}
	if proxy != nil {
		net.client.Transport = &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return net, nil
}

func (net *PwnNetworking) Close() {
	net.client.CloseIdleConnections()
}

func (net PwnNetworking) genAuthToken() string {
	token := append([]string{}, "myciceroapp", net.uUID, fmt.Sprint(time.Now().UnixNano()/1000000))

	hasher := hmac.New(sha1.New, []byte{99, 104, 105, 97, 118, 101, 100, 105, 116, 101, 115, 116, 51, 51, 51, 51})
	hasher.Write([]byte(strings.Join(token, "")))
	token = append(token, base64.StdEncoding.EncodeToString(hasher.Sum(nil)))

	return strings.Join(token, "\\")
}

func (net *PwnNetworking) post(method string, postArg map[string]interface{}, needID ...bool) ([]byte, error) {
	if len(needID) == 0 || needID[0] == true {
		postArg["IdDevice"] = "71d779f052da76c7"
	}

	postArg["SecurityToken"] = "MDgtMTEtMjAxOSAwMTo0MTE2NzY="
	reqBody, err := json.Marshal(postArg)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://www.autobus.it/proxy.imomo/proxy.ashx?url=tsgw.v1se/json/"+method+"&_t_nocache="+fmt.Sprint(time.Now().UnixNano()/1000000),
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return []byte{}, err
	}

	for k, v := range map[string]string{
		"auth":            net.genAuthToken(),
		"culture":         "en-GB",
		"timeIsSynched":   "true",
		"Accept-Language": "en-GB, en;q=0.9, it-IT;q=0.8, it;q=0.7, de-DE;q=0.8, de;q=0.7, fr-FR;q=0.8, fr;q=0.7, ru-RU;q=0.8, ru;q=0.7, tr-TR;q=0.8, tr;q=0.7",
		"RequestNFC":      "0",
		"Client":          "myCicero;6.6.2",
		"deviceID":        net.devID,
		"deviceInfo":      "Google; Android SDK built for x86; Android 9",
		"Accept":          "application/json",
		"Content-Type":    "application/json",
		"User-Agent":      "Dalvik/2.1.0 (Linux; U; Android 9; Android SDK built for x86 Build/PSR1.180720.075)",
		"Host":            "www.autobus.it",
		"Connection":      "Keep-Alive",
		"Accept-Encoding": "gzip",
		"Content-Length":  fmt.Sprint(len(reqBody)),
	} {
		req.Header[k] = []string{v}
	}
	res, err := net.client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	resBody := make([]byte, 0)
	if res.Header.Get("Content-Encoding") == "gzip" {
		decompresser, err := gzip.NewReader(res.Body)
		if err != nil {
			return []byte{}, err
		}

		resBody, err = ioutil.ReadAll(decompresser)
		if err != nil {
			return []byte{}, err
		}

		if err := decompresser.Close(); err != nil {
			return []byte{}, err
		}
	} else {
		resBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return []byte{}, err
		}
	}

	if err := res.Body.Close(); err != nil {
		return []byte{}, err
	}
	return resBody, nil
}

func (net PwnNetworking) Search(searchString string) ([]byte, error) {
	return net.post(
		"Search",
		map[string]interface{}{
			"CathegoryRows": 4,
			"SearchString":  strings.ToUpper(searchString),
			"Tipi":          [1]int{1},
			"Punto":         nil,
			"Ambito":        nil,
		},
		false,
	)

	/*var ret struct {
		Oggetti []Location
	}
	if err = json.Unmarshal(jobj, &ret); err != nil {
		return []Location{}, err
	}

	return ret.Oggetti, nil*/
}

func (net PwnNetworking) FindTPSolutions(
	leavingTime time.Time,
	modeFilter ModeFilter,
	origin, destination Location,
) ([]byte, error) {
	leavingDate := time.Date(
		leavingTime.Year(),
		leavingTime.Month(),
		leavingTime.Day(),
		0, 0, 0, 0,
		leavingTime.Location(),
	)

	return net.post(
		"FindTPSolutions",
		map[string]interface{}{
			"Ambiente": map[string][]int{
				"Ambiti": []int{0, 1, 2},
			},
			"DataPartenza":            "/Date(" + fmt.Sprint(leavingDate.UnixNano()/time.Millisecond.Nanoseconds()) + "+0200)/",
			"FiltroModalita":          modeFilter,
			"Intermodale":             false,
			"IstatComuneDestinazione": destination.Istat,
			"IstatComuneOrigine":      origin.Istat,
			"DescrizionePartenza":     origin.Description,
			"DescrizioneDestinazione": destination.Description,
			"MezzoTrasporto":          1,
			"ModalitaRicerca":         0,
			"OraDa":                   "/Date(" + fmt.Sprint(leavingTime.UnixNano()/time.Millisecond.Nanoseconds()) + "+0200)/",
			"PuntoDestinazione":       destination.Point(),
			"PuntoOrigine":            origin.Point(),
			"TPLogDetails": map[string]interface{}{
				"ComuneArrivo":      "",
				"ComunePartenza":    "",
				"IPAddress":         nil,
				"IndirizzoArrivo":   "",
				"IndirizzoPartenza": "",
			},
			"TipoPercorso": 0,
		},
	)
}

func (net PwnNetworking) GetTPSolutionDetails(contextID string, solNum int) ([]byte, error) {
	return net.post(
		"GetTPSolutionDetail",
		map[string]interface{}{
			"IdContesto":      contextID,
			"Lingua":          "de",
			"NumeroSoluzione": solNum,
		},
	)
}

func main() {
	var proxy *url.URL
	//proxy, err := url.Parse("http://127.0.0.1:8080")

	net, err := NewNet("0f003f82-c0d5-11e9-917d-fc4596ef1588", proxy)
	defer net.Close()
	if err != nil {
		fmt.Println("[pwnNetwork::New] " + err.Error())
		return
	}

	/*body, err := net.GetTPSolutionDetails(
		"0#InfoUtenza#640699#55#640699001|",
		640699001,
	)
	if err != nil {
		fmt.Println("[pwnNetwork::FindTPSolutions] " + err.Error())
		return
	}

	fmt.Println(string(body))*/

	net.Close()
}
