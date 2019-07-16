package lash_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
)

func Test_http_requests(t *testing.T) {

	t.Run("default method is GET", func(t *testing.T) {
		var method string
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
		})
		defer ts.Close()

		resp := lash.
			Curl(ts.URL + "/any").
			Response()

		assert.NotNil(t, resp)
		assert.NoError(t, lash.DefaultSession.Err())
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "GET", method)
	})
	t.Run("body can be read as string or byte slice", func(t *testing.T) {
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`some content`))
		})
		defer ts.Close()

		session := lash.NewSession()
		resp := session.
			Curl(ts.URL + "/any").
			Response()

		assert.NotNil(t, resp)
		assert.NoError(t, session.Err())
		assert.Equal(t, "some content", resp.BodyString())
		assert.Equal(t, []byte("some content"), resp.BodyBytes())
	})
	t.Run("body can be read as json to struct", func(t *testing.T) {
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"name":"any-name","count":42}`))
		})
		defer ts.Close()

		var actual struct {
			Name  string
			Count int
		}
		session := lash.NewSession()
		session.
			Curl(ts.URL + "/any").
			Response().
			FromJSON(&actual)

		require.NoError(t, session.Err())
		assert.Equal(t, "any-name", actual.Name)
		assert.Equal(t, 42, actual.Count)
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
		var currentStatus int
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(currentStatus)
		})
		defer ts.Close()

		for _, td := range testData {
			currentStatus = td.status

			resp := lash.
				Curl(ts.URL+"/any").
				AllowResponses(200, 404).
				Response()

			assert.Equal(t, td.isError, lash.DefaultSession.IsError(), td.name)
			assert.Equal(t, td.isError, resp.IsError(), td.name)
		}
	})
	t.Run("can send a body", func(t *testing.T) {
		var actualBody []byte
		var method string
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			buf, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			actualBody = buf
			defer func() { _ = r.Body.Close() }()
		})
		defer ts.Close()

		t.Run("with PUT", func(t *testing.T) {
			session := lash.NewSession()
			resp := session.
				Curl(ts.URL + "/any").
				Put([]byte("some content")).
				Response()

			require.NoError(t, session.Err())

			assert.Equal(t, "PUT", method)
			assert.Equal(t, "some content", string(actualBody))
			assert.Equal(t, 200, resp.StatusCode())
		})
		t.Run("with POST", func(t *testing.T) {
			session := lash.NewSession()
			resp := session.
				Curl(ts.URL + "/any").
				Post([]byte("more content")).
				Response()

			require.NoError(t, session.Err())

			assert.Equal(t, "POST", method)
			assert.Equal(t, "more content", string(actualBody))
			assert.Equal(t, 200, resp.StatusCode())
		})
	})
}

func makeTestServer(handler http.HandlerFunc) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	return ts
}
