package tasks

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"

	//	@name	render
	"github.com/timohahaa/transcoder/internal/utils/render"
	"github.com/timohahaa/transcoder/pkg/validate"
)

// @Summary	Create task
// @Tags		Tasks
// @Param		CreateParams	body		task.CreateForm		true	"Create Params"
// @Success	200				{object}	task.Task			"Task Model"
// @Failure	default			{object}	render.HTTPError	"Error"
// @Router		/v1/tasks/ [post]
func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		form task.CreateForm
	)

	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		render.Error(w, err)
		return
	}

	if invParams := validate.Struct(&form); len(invParams) != 0 {
		render.Error(w, invParams)
		return
	}

	task, err := h.mod.task.Create(ctx, form)
	if err != nil {
		render.Error(w, err)
		return
	}

	render.JSON(w, task)
}

// @Summary	Get task
// @Tags		Tasks
// @Param		task_id	path		string				true	"Task ID"
// @Success	200		{object}	task.Task			"Task Model"
// @Failure	default	{object}	render.HTTPError	"Error"
// @Router		/v1/tasks/{task_id}/ [get]
func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		id, err = uuid.Parse(chi.URLParam(r, "task_id"))
	)
	if err != nil {
		render.Error(w, err)
		return
	}

	task, err := h.mod.task.Get(ctx, id)
	if err != nil {
		render.Error(w, err)
		return
	}

	render.JSON(w, task)
}

// @Summary	Delete task
// @Tags		Tasks
// @Param		task_id	path		string	true	"Task ID"
// @Success	200		{object}	nil		"Response"
// @Failure	default	{object}	render.HTTPError	"Error"
// @Router		/v1/tasks/{task_id}/ [delete]
func (h *handlers) delete(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		id, err = uuid.Parse(chi.URLParam(r, "task_id"))
	)
	if err != nil {
		render.Error(w, err)
		return
	}

	if err := h.mod.queue.SkipTask(ctx, id); err != nil {
		render.Error(w, err)
		return
	}

	if err := h.mod.task.Delete(ctx, id); err != nil {
		render.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary	Cancel task
// @Tags		Tasks
// @Param		task_id	path		string	true	"Task ID"
// @Success	200		{object}	nil		"Response"
// @Failure	default	{object}	render.HTTPError	"Error"
// @Router		/v1/tasks/{task_id}/cancel/ [post]
func (h *handlers) cancel(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		id, err = uuid.Parse(chi.URLParam(r, "task_id"))
	)
	if err != nil {
		render.Error(w, err)
		return
	}

	if err := h.mod.queue.SkipTask(ctx, id); err != nil {
		render.Error(w, err)
		return
	}

	if err := h.mod.task.UpdateStatus(ctx, id, task.StatusCanceled, nil); err != nil {
		render.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type taskProgress struct {
	Progress int64 `json:"progress"`
}

// @Summary	Get task progress
// @Tags		Tasks
// @Param		task_id	path		string				true	"Task ID"
// @Success	200		{object}	taskProgress			"Task progress"
// @Failure	default	{object}	render.HTTPError	"Error"
// @Router		/v1/tasks/{task_id}/progress/ [get]
func (h *handlers) progress(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = r.Context()
		id, err = uuid.Parse(chi.URLParam(r, "task_id"))
	)
	if err != nil {
		render.Error(w, err)
		return
	}

	prog, err := h.mod.task.GetProgress(ctx, id)
	if err != nil {
		render.Error(w, err)
		return
	}

	render.JSON(w, taskProgress{Progress: prog})
}
