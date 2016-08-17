package db

import (
	"testing"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

var pinged bool
var closed bool

type fakeClient struct {
}

func (c *fakeClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	pinged = true
	return time.Duration(3), "", nil
}

func (c *fakeClient) Close() error {
	closed = true
	return nil
}

func (c *fakeClient) Write(bp client.BatchPoints) error {
	return nil
}

func (c *fakeClient) Query(q client.Query) (*client.Response, error) {
	return &client.Response{}, nil
}

func fakeInfluxdbClient() (client.Client, error) {
	return &fakeClient{}, nil
}

func TestQuery(t *testing.T) {
	originFunc := CreateInfluxHttpClient
	CreateInfluxHttpClient = fakeInfluxdbClient
	defer func() {
		CreateInfluxHttpClient = originFunc
	}()

	_, err := Query("show measurements")
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if closed != true {
		t.Error("unexpected test result: conn should be closed.")
	}
	// clean up
	closed = false
}
