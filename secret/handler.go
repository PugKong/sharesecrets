package secret

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pugkong/sharesecrets/html"
)

type Handler struct {
	secrets  *Service
	renderer *html.Renderer
}

func NewHandler(secrets *Service, renderer *html.Renderer) *Handler {
	return &Handler{
		secrets:  secrets,
		renderer: renderer,
	}
}

func (h *Handler) Share(w http.ResponseWriter, r *http.Request) {
	const attempts = 3

	data := createData{
		Passphrase: r.Form.Get("passphrase"),
		Message:    r.Form.Get("message"),
		Expire: createExpireData{
			Amount: cmp.Or(r.Form.Get("expire_amount"), "15"),
			Unit:   cmp.Or(r.Form.Get("expire_unit"), "minutes"),
		},
	}

	if r.Method == http.MethodPost {
		data.Violations = h.validateShareData(data)
		if len(data.Violations) > 0 {
			h.renderer.Component(r.Context(), w, http.StatusOK, createPage(data))

			return
		}

		request := StoreRequest{
			Passphrase: data.Passphrase,
			Message:    data.Message,
			Attempts:   attempts,
			ExpireAt:   time.Now().Add(data.Expire.Duration()),
		}

		secretID, err := h.secrets.Store(r.Context(), request)
		if err == nil {
			secretURL := fmt.Sprintf("%s/%s", r.Header.Get("origin"), secretID)
			page := sharePage(secretURL)

			h.renderer.Component(r.Context(), w, http.StatusOK, page)

			return
		}

		h.renderer.ServerError(r.Context(), w, err)

		return
	}

	h.renderer.Component(r.Context(), w, http.StatusOK, createPage(data))
}

func (h *Handler) validateShareData(request createData) []string {
	var violations []string

	const maxPassphraseLen = 32
	if len(request.Passphrase) > maxPassphraseLen {
		violations = append(violations, "The passphrase must be less than or equal to 32 bytes")
	}

	const maxMessageLen = 4 * 1024
	if len(request.Message) > maxMessageLen {
		violations = append(violations, "The message must be less than or equal to 4 kilobytes")
	}

	if request.Expire.Duration() <= 0 {
		violations = append(violations, "The expire field must be positive")
	}

	if request.Expire.Duration() > 24*time.Hour {
		violations = append(violations, "Expire must be less than 1 day")
	}

	return violations
}

func (h *Handler) Open(w http.ResponseWriter, r *http.Request) {
	data := openData{Passphrase: r.Form.Get("passphrase")}

	if r.Method == http.MethodPost {
		request := RetrieveRequest{
			Key:        r.PathValue("key"),
			Passphrase: data.Passphrase,
		}

		message, err := h.secrets.Retrieve(r.Context(), request)
		if err == nil {
			page := viewPage(message)

			h.renderer.Component(r.Context(), w, http.StatusOK, page)

			return
		}

		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrExpired) || errors.Is(err, ErrInvalidPassphrase) {
			data.Violations = append(data.Violations, "Message not found or invalid passphrase")
		} else {
			h.renderer.ServerError(r.Context(), w, err)

			return
		}
	}

	h.renderer.Component(r.Context(), w, http.StatusOK, openPage(data))
}
