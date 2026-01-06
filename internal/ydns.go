package ydns

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RequestInfo struct {
	Family   string
	Base     string
	RecordID string
	Host     string
	User     string
	Pass     string
	IP       string
}

// Run will run the update operation.
// Run will execute the update operation. The `family` parameter can be
// "ipv4", "ipv6" or any other value (means no preference).
func Run(requestInfo *RequestInfo) error {
	u, err := url.Parse(requestInfo.Base)
	if err != nil {
		return errors.Wrap(err, "cannot create url")
	}

	values := url.Values{}
	values.Set("host", requestInfo.Host)

	if requestInfo.IP != "" {
		values.Set("ip", requestInfo.IP)
	}

	if requestInfo.RecordID != "" {
		values.Set("record_id", requestInfo.RecordID)
	}

	u.RawQuery = values.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "cannot create request")
	}

	req.SetBasicAuth(requestInfo.User, requestInfo.Pass)

	// ignore user and pass
	requestInfo.User = "***"
	requestInfo.Pass = "***"
	logrus.WithField("requestInfo", fmt.Sprintf("%#v", *requestInfo)).Info("updating record")

	// Build an HTTP client that can be forced to use IPv4 or IPv6 when needed.
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return errors.New("cannot get default transport")
	}
	tr := transport.Clone()

	switch requestInfo.Family {
	case "ipv4", "4":
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, "tcp4", addr)
		}
	case "ipv6", "6":
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, "tcp6", addr)
		}
	}

	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "cannot perform http get")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read the body")
	}

	logrus.WithFields(logrus.Fields{
		"body":   string(body),
		"status": res.StatusCode,
	}).Debug("got response from api")

	// Log based on request status code
	switch res.StatusCode {
	case http.StatusOK:
		logrus.WithField("host", requestInfo.Host).Info("update was successful")
	case http.StatusBadRequest:
		return errors.New("failed to perform request due to invalid input parameters")
	case http.StatusUnauthorized:
		return errors.New("failed to perform request due to authentication issues")
	case http.StatusNotFound:
		return errors.New("failed to perform request because the host you'd like to update cannot be found")
	default:
		return errors.Errorf("some unknown error occurred: %s", res.Status)
	}

	return nil
}
