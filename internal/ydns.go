package ydns

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Run will run the update operation.
// Run will execute the update operation. The `family` parameter can be
// "ipv4", "ipv6" or any other value (means no preference).
func Run(base, host, ip, record_id, user, pass, family string) error {
	u, err := url.Parse(base)
	if err != nil {
		return errors.Wrap(err, "cannot create url")
	}

	values := url.Values{}
	values.Set("host", host)

	if ip != "" {
		logrus.WithField("ip", ip).Info("updating record")
		values.Set("ip", ip)
	}

	if record_id != "" {
		logrus.WithField("record_id", record_id).Info("updating record")
		values.Set("record_id", record_id)
	}

	u.RawQuery = values.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "cannot create request")
	}

	req.SetBasicAuth(user, pass)

	logrus.WithField("host", host).Info("updating record")

	// Build an HTTP client that can be forced to use IPv4 or IPv6 when needed.
	transport, _ := http.DefaultTransport.(*http.Transport)
	tr := transport.Clone()

	switch family {
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
		logrus.WithField("host", host).Info("update was successful")
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
