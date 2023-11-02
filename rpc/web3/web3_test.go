package web3

import (
	"net/http"
	"testing"
	"time"
)

func TestRateLimitMiddleware(t *testing.T) {
	url := "http://localhost/v1/getBlock?number=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			t.Logf("Request %d: Rate limit exceeded", i+1)
		} else if resp.StatusCode == http.StatusOK {
			t.Logf("Request %d: OK", i+1)
		} else {
			t.Errorf("Request %d: Unexpected status code %d", i+1, resp.StatusCode)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
