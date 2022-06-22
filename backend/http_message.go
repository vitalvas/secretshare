package backend

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

const (
	expDefault = 10800
	expMax     = 604800
)

func (backend *Backend) httpSaveMessage(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Message string
		Exp     int
		Pin     string
	}{}

	if err := render.DecodeJSON(r.Body, &request); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	if len(request.Pin) > 0 {
		if _, err := strconv.ParseUint(request.Pin, 10, 32); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, JSON{"error": "error validate pin"})
			return
		}
	} else {
		request.Pin = "0"
	}

	if request.Exp == 0 {
		request.Exp = expDefault
	} else if request.Exp > expMax {
		request.Exp = expMax
	}

	msg, err := backend.makeMessage(time.Second*time.Duration(request.Exp), request.Message, request.Pin)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, JSON{"key": msg.Key, "exp": msg.Exp.Format(time.RFC3339)})
}

func (backend *Backend) httpGetMessage(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	pin := chi.URLParam(r, "pin")

	if pin == "" {
		pin = "0"
	}

	if key == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": "no key passed"})
	}

	msg, err := backend.loadMessage(key, pin)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			render.Status(r, http.StatusNotFound)

		case ErrFailedToDecrypt:
			render.Status(r, http.StatusExpectationFailed)

		default:
			render.Status(r, http.StatusInternalServerError)
		}

		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	render.JSON(w, r, JSON{"key": key, "message": string(msg.Data)})
}
