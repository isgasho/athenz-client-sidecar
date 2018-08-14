package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"ghe.corp.yahoo.co.jp/athenz/hcc-k8s/config"
	"github.com/pkg/errors"
)

type UDB interface {
	GetByGUID(guid string) (map[string]string, error)
}

type udb struct {
	yca   YCA
	host  string
	keys  string
	appID string
}

func NewUDBClient(cfg config.UDB, yca YCA) UDB {
	return &udb{
		yca:   yca,
		host:  fmt.Sprintf("%s://%s:%d/%s/%s", cfg.Scheme, cfg.Host, cfg.Port, cfg.Version, "users"),
		appID: cfg.AppID,
		keys:  strings.Join(cfg.Keys, ","),
	}
}

//GetByGUID get data by GUID
func (u *udb) GetByGUID(guid string) (map[string]string, error) {
	return u.doRequest(http.MethodGet,
		fmt.Sprintf("%s/%s?fields=%s", u.host, guid, u.keys),
		"",
		nil)
}

func (u *udb) doRequest(method, url, cookie string, body io.Reader) (map[string]string, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	cert, err := u.yca.GetCertificate(u.appID)
	if err != nil {
		return nil, err
	}

	req.Header.Del("Yahoo-App-Auth")
	req.Header.Set("Yahoo-App-Auth", cert)

	req.Header.Del("Content-Type")
	req.Header.Set("Content-Type", "application/json")

	if len(cookie) > 0 {
		req.Header.Del("Cookie")
		req.Header.Set("Cookie", cookie)
	}

	// TODO 別のクライアントにする
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	// StatusOK 200 でリクエスト成功
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Error: response status " + strconv.Itoa(res.StatusCode))
	}

	var data map[string]string
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	var b []byte
	for k, v := range data {
		b, err = base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		data[k] = string(b[:len(b)])
		b = b[:0]
	}

	return data, nil
}
