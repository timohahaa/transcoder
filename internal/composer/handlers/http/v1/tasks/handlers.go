package tasks

import "net/http"

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {}

func (h *handlers) delete(w http.ResponseWriter, r *http.Request) {}

func (h *handlers) cancel(w http.ResponseWriter, r *http.Request) {}

func (h *handlers) progress(w http.ResponseWriter, r *http.Request) {}
