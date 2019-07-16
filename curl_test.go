package lash_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
)

func Test_http_requests(t *testing.T) {
	var handler http.HandlerFunc
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	t.Run("default method is GET", func(t *testing.T) {
		handler = func(w http.ResponseWriter, r *http.Request) {}
		resp := lash.
			Curl(ts.URL + "/any").
			Response()

		assert.NotNil(t, resp)
		assert.NoError(t, lash.DefaultSession.Err())
		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})
	t.Run("body can be read as string or byte slice", func(t *testing.T) {

		handler = func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`some content`))
		}
		resp := lash.
			Curl(ts.URL + "/any").
			Response()

		assert.NotNil(t, resp)
		assert.NoError(t, lash.DefaultSession.Err())
		assert.Equal(t, "some content", resp.BodyString())
		assert.Equal(t, []byte("some content"), resp.BodyBytes())
	})
	t.Run("can limit which statuses count as error", func(t *testing.T) {

		testData := []struct {
			name    string
			status  int
			isError bool
		}{
			{"200 in the list", 200, false},
			{"404 in the list", 404, false},
			{"418 is not in the list", 418, true},
		}
		for _, td := range testData {
			handler = func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(td.status)
				_, _ = w.Write([]byte(`some content`))

			}

			resp := lash.
				Curl(ts.URL+"/any").
				AllowResponses(200, 404).
				Response()

			assert.Equal(t, td.isError, lash.DefaultSession.IsError(), td.name)
			assert.Equal(t, td.isError, resp.IsError(), td.name)
		}
	})

}
