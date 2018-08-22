package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"ghe.corp.yahoo.co.jp/athenz/hcc-k8s/config"
	"github.com/kpango/glg"
	"github.com/pkg/errors"
)

type HCC interface {
	StartCertUpdater(ctx context.Context)
	GetCertProvider() CertProvider
}

type hcc struct {
	certs         sync.Map
	ip            string
	hostname      string
	token         TokenProvider
	nextExpire    time.Time
	lastRefreshed time.Time
}

type certificate struct {
	AppID string `xml:"appid,attr"`
	Cert  string `xml:",chardata"`
}

type certificates struct {
	Hostname     string        `xml:"hostname,attr"`
	Certificates []certificate `xml:"certificate"`
}

type CertProvider func(string) (string, error)

const (
	zts                   = "zts.athenz.yahoo.co.jp:4443/wsca/v1"
	defaultCertExpireTime = 30 * time.Minute // sentinel for when no certs are returned
)

var (
	ErrCertNotFound = errors.New("certification not found")
)

func NewHCC(cfg config.HCC, prov TokenProvider) (HCC, error) {
	return &hcc{
		certs:         sync.Map{},
		ip:            cfg.IP,
		hostname:      cfg.Hostname,
		token:         prov,
		nextExpire:    time.Now(),
		lastRefreshed: time.Now(),
	}, nil
}

func (h *hcc) GetCertProvider() CertProvider {
	return h.getCertificate
}

func (h *hcc) getCertificate(appID string) (string, error) {
	cert, ok := h.certs.Load(appID)
	if !ok {
		return "", ErrCertNotFound
	}

	return cert.(string), nil
}

func (h *hcc) update() error {
	u := fmt.Sprintf("https://%s/caontainercerts/mh/%s?d=%d&ip=%s", zts, h.hostname, time.Hour/time.Second, url.QueryEscape(h.ip))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	token, err := h.token()
	if err != nil {
		return err
	}
	req.Header.Set("Yahoo-Principal-Auth", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if res.Body != nil {
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned status code %d", u, res.StatusCode)
	}

	var certs certificates

	err = xml.NewDecoder(res.Body).Decode(&certs)
	if err != nil {
		return err
	}

	sentinel := time.Now().Add(365 * 24 * time.Hour)
	earliestExpiry := sentinel
	for _, cert := range certs.Certificates {
		exp, err := h.checkExpire(cert.Cert)
		if err != nil {
			continue
		}
		if exp.Before(earliestExpiry) {
			earliestExpiry = exp
		}
		h.certs.Store(cert.AppID, cert.Cert)
	}

	if earliestExpiry != sentinel {
		h.nextExpire = earliestExpiry
	} else {
		h.nextExpire = time.Now().Add(defaultCertExpireTime)
	}

	h.lastRefreshed = time.Now()

	return nil
}

func (h *hcc) checkExpire(cert string) (time.Time, error) {
	for _, part := range strings.Split(cert, ";") {
		if strings.HasPrefix(part, "k=") {
			v, err := strconv.ParseInt(strings.TrimPrefix(part, "k="), 10, 64)
			if err != nil {
				return time.Time{}, err
			}
			return time.Unix(v, 0), nil
		}
	}
	return time.Time{}, nil
}

func (h *hcc) StartCertUpdater(ctx context.Context) {
	go func() {
		var err error
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if time.Now().After(h.nextExpire) {
					err = h.update()
					if err != nil {
						glg.Error(err)
					}
				}
			}
		}
	}()
}
