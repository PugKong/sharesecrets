package html

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pugkong/sharesecrets/loggertest"
	"github.com/stretchr/testify/require"
)

func TestRecoverMiddleware(t *testing.T) {
	assets, err := MakeAssets()
	if err != nil {
		t.Fatal(err)
	}

	renderContext := context.Background()
	renderContext = context.WithValue(renderContext, assetsKey, assets)
	renderContext = context.WithValue(renderContext, csrfKey, "token")

	t.Run("it recovers panic and logs it", func(t *testing.T) {
		logger, logs := loggertest.New()
		renderer := NewRenderer(logger)

		middleware := NewRecoverMiddleware(logger, renderer)
		handler := middleware.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("some panic")
		}))

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(renderContext)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		response := recorder.Result()
		defer response.Body.Close()

		require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		require.Contains(t, recorder.Body.String(), "500: Something broke on our side")

		// here we have two log entries: one from middleware and one from Renderer.ServerError
		out := logs(t)
		require.Len(t, out, 2)
		require.Equal(t, "ERROR", out[0]["level"])
		require.Equal(t, "Panic recovered", out[0]["msg"])
		require.Equal(t, "some panic", out[0]["error"])
		require.NotEmpty(t, out[0]["stack"])
	})
}
