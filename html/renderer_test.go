package html

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pugkong/sharesecrets/loggertest"
	"github.com/stretchr/testify/require"
)

func TestRenderer(t *testing.T) {
	assets, err := MakeAssets()
	if err != nil {
		t.Fatal(err)
	}

	renderContext := context.Background()
	renderContext = context.WithValue(renderContext, assetsKey, assets)
	renderContext = context.WithValue(renderContext, csrfKey, "token")

	tests := map[string]struct {
		render     func(*Renderer, http.ResponseWriter)
		statusCode int
		contains   string
		logs       []map[string]any
	}{
		"it renders component with 200 status code": {
			render: func(r *Renderer, w http.ResponseWriter) {
				r.Component(renderContext, w, http.StatusOK, Layout("200 page"))
			},
			statusCode: http.StatusOK,
			contains:   "200 page",
		},
		"it renders component with 400 statuc code": {
			render: func(r *Renderer, w http.ResponseWriter) {
				r.Component(renderContext, w, http.StatusBadRequest, Layout("400 page"))
			},
			statusCode: http.StatusBadRequest,
			contains:   "400 page",
		},
		"it logs component render error": {
			render: func(r *Renderer, w http.ResponseWriter) {
				r.Component(context.Background(), w, http.StatusOK, Layout("200 page"))
			},
			statusCode: http.StatusOK,
			logs: []map[string]any{
				{
					"level": "ERROR",
					"msg":   "Failed to render component",
					"error": Layout("").Render(context.Background(), io.Discard).Error(),
				},
			},
		},
		"it renders user error page and logs user error": {
			render:     func(r *Renderer, w http.ResponseWriter) { r.UserError(renderContext, w, io.EOF) },
			statusCode: http.StatusBadRequest,
			contains:   "400: Something broke on your side",
			logs:       []map[string]any{{"level": "INFO", "msg": "User input error", "error": io.EOF.Error()}},
		},
		"it logs render error of user error page": {
			render:     func(r *Renderer, w http.ResponseWriter) { r.UserError(context.Background(), w, io.EOF) },
			statusCode: http.StatusBadRequest,
			logs: []map[string]any{
				{"level": "INFO", "msg": "User input error", "error": io.EOF.Error()},
				{
					"level": "ERROR",
					"msg":   "Failed to render user error page",
					"error": UserError().Render(context.Background(), io.Discard).Error(),
				},
			},
		},
		"it renders server error page and logs server error": {
			render:     func(r *Renderer, w http.ResponseWriter) { r.ServerError(renderContext, w, io.EOF) },
			statusCode: http.StatusInternalServerError,
			contains:   "500: Something broke on our side",
			logs:       []map[string]any{{"level": "ERROR", "msg": "Server error", "error": io.EOF.Error()}},
		},
		"it logs render error of server error page": {
			render:     func(r *Renderer, w http.ResponseWriter) { r.ServerError(context.Background(), w, io.EOF) },
			statusCode: http.StatusInternalServerError,
			contains:   "",
			logs: []map[string]any{
				{"level": "ERROR", "msg": "Server error", "error": io.EOF.Error()},
				{
					"level": "ERROR",
					"msg":   "Failed to render server error page",
					"error": ServerError().Render(context.Background(), io.Discard).Error(),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			logger, logs := loggertest.New()
			renderer := NewRenderer(logger)
			recorder := httptest.NewRecorder()

			test.render(renderer, recorder)

			response := recorder.Result()
			defer response.Body.Close()

			require.Equal(t, test.statusCode, response.StatusCode)
			require.Contains(t, recorder.Body.String(), test.contains)
			require.Equal(t, test.logs, logs(t))
		})
	}
}
