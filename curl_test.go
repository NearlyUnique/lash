package lash_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_http_requests(t *testing.T) {

	t.Run("default request method is GET", func(t *testing.T) {
		var method string
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
		})
		defer ts.Close()

		scope := lash.NewScope()
		resp := scope.
			Curl(ts.URL + "/any").
			Response()

		assert.NotNil(t, resp)
		assert.NoError(t, scope.Err())
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "GET", method)
	})
	t.Run("response body can be read as string or byte slice", func(t *testing.T) {
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`some content`))
		})
		defer ts.Close()

		scope := lash.NewScope()
		resp := scope.
			Curl(ts.URL + "/any").
			Response()

		assert.NotNil(t, resp)
		assert.NoError(t, scope.Err())
		assert.Equal(t, "some content", resp.BodyString())
		assert.Equal(t, []byte("some content"), resp.BodyBytes())
	})
	t.Run("response body can be read as json to struct", func(t *testing.T) {
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"name":"any-name","count":42}`))
		})
		defer ts.Close()

		var actual struct {
			Name  string
			Count int
		}
		scope := lash.NewScope()
		scope.
			Curl(ts.URL + "/any").
			Response().
			FromJSON(&actual)

		require.NoError(t, scope.Err())
		assert.Equal(t, "any-name", actual.Name)
		assert.Equal(t, 42, actual.Count)
	})
	t.Run("can limit which http response statuses count as error", func(t *testing.T) {
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

			scope := lash.NewScope()
			scope.OnError(lash.Ignore)
			resp := scope.
				Curl(ts.URL+"/any").
				AllowResponses(200, 404).
				Response()

			assert.Equal(t, td.isError, scope.IsError(), td.name)
			assert.Equal(t, td.isError, resp.IsError(), td.name)
			assert.Equal(t, td.status, resp.StatusCode())
		}
	})
	t.Run("can allow all http status codes to be non-error", func(t *testing.T) {
		testData := []int{
			200, 201, 400, 404, 500, 501,
		}
		var currentStatus int
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(currentStatus)
		})
		defer ts.Close()
		scope := lash.NewScope()
		for _, status := range testData {
			currentStatus = status

			resp := scope.
				Curl(ts.URL + "/any").
				AllowResponses(lash.AnyHTTPStatus).
				Response()

			assert.NoError(t, scope.Err())
			assert.Equal(t, status, resp.StatusCode())
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
			scope := lash.NewScope()
			resp := scope.
				Curl(ts.URL + "/any").
				Put([]byte("some content")).
				Response()

			require.NoError(t, scope.Err())

			assert.Equal(t, "PUT", method)
			assert.Equal(t, "some content", string(actualBody))
			assert.Equal(t, 200, resp.StatusCode())
		})
		t.Run("with POST", func(t *testing.T) {
			scope := lash.NewScope()
			resp := scope.
				Curl(ts.URL + "/any").
				Post([]byte("more content")).
				Response()

			require.NoError(t, scope.Err())

			assert.Equal(t, "POST", method)
			assert.Equal(t, "more content", string(actualBody))
			assert.Equal(t, 200, resp.StatusCode())
		})
	})
	t.Run("request can contain any header", func(t *testing.T) {
		called := 0
		ts := makeTestServer(func(w http.ResponseWriter, r *http.Request) {
			called++

			// note when accessing the map directly the canonical casing is required
			assert.Equal(t, "actual-value", r.Header.Get("header-name"))
			assert.Equal(t, 1, len(r.Header["Header-Name"]))

			assert.Equal(t, 2, len(r.Header["Header-Set"]))
			assert.Contains(t, r.Header["Header-Set"], "one")
			assert.Contains(t, r.Header["Header-Set"], "two")

		})

		scope := lash.NewScope()
		_ = scope.
			Curl(ts.URL+"/any").
			Header("header-name", "this will get overridden").
			Header("header-name", "actual-value").
			Header("header-set", "one").
			AddHeader("header-set", "two").
			Response()

		assert.Equal(t, 1, called)
		require.NoError(t, scope.Err())
	})
	t.Run("can make a func to add common settings", func(t *testing.T) {
		scope := lash.NewScope()

		custom := func(r *lash.HTTPRequest) {
			r.Header("key1", "value1").
				Header("key2", "value2").
				AllowResponses(200)
		}

		request := scope.
			Curl("https://example.com").
			CommonFunc(custom).
			Header("key0", "value0")

		assert.Equal(t, "value0", request.Req.Header.Get("Key0"))
		assert.Equal(t, "value1", request.Req.Header.Get("Key1"))
		assert.Equal(t, "value2", request.Req.Header.Get("Key2"))
		require.NoError(t, scope.Err())
	})
	t.Run("env vars can be in the url", func(t *testing.T) {
		scope := lash.NewScope()

		require.NoError(t, os.Setenv("url_env_test", "example.com"))
		request := scope.Curl("https://$url_env_test/one/$0", 2)

		assert.Equal(t, "https://example.com/one/2", request.Req.URL.String())
		require.NoError(t, scope.Err())
	})
	t.Run("env vars can be in the header values", func(t *testing.T) {
		scope := lash.NewScope()

		require.NoError(t, os.Setenv("h1", "value one"))
		require.NoError(t, os.Setenv("h2_2", "second"))

		request := scope.Curl("https://example.com").
			Header("h1", "$h1").
			Header("h2", "x-$0 $1", "one", 2).
			AddHeader("h2", "$h2_2:$0", "any")

		assert.Equal(t, "value one", request.Req.Header.Get("H1"))
		assert.Contains(t, request.Req.Header["H2"], "x-one 2")
		assert.Contains(t, request.Req.Header["H2"], "second:any")
		require.NoError(t, scope.Err())
	})

	//t.Run("well known headers have helper functions", func(t *testing.T) {
	//	scope := lash.NewScope()
	//	request := scope.
	//		Curl("http://example.com").
	//		Header("", "")
	//		//Authorization(lash.AuthzBearer, "some-token").
	//		//ContentType(lash.ApplicationJSON).
	//		//Accept(lash.ApplicationJSON).
	//		//AcceptLang(language.English).
	//		//UserAgent("some-user-agent")
	//
	//	assert.Equal(t, "Bearer some-token", request.Req.Header.Get("authorization"))
	//	assert.Equal(t, "application/json", request.Req.Header.Get("content-type"))
	//	assert.Equal(t, "application/json", request.Req.Header.Get("accept"))
	//	assert.Equal(t, "some-user-agent", request.Req.Header.Get("user-agent"))
	//})
}

func makeTestServer(handler http.HandlerFunc) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	return ts
}
