package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (h *handlers) getFile(w http.ResponseWriter, r *http.Request) {
	var (
		q          = r.URL.Query()
		sourcePath = q.Get("filepath")
		l          = log.WithFields(log.Fields{
			"mod":    "http",
			"source": sourcePath,
		})
	)

	f, err := os.Open(sourcePath)
	if err != nil {
		l.Error(err)
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if s.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", s.Size()))

	n, err := io.Copy(w, f)
	if err != nil {
		l.Errorf("copy file: %v", err)
	}
	if n != s.Size() {
		l.Error("resp body size mismatch: filesize = %v, body = %v", s.Size(), n)
	}
}

func (h *handlers) getChunk(w http.ResponseWriter, r *http.Request) { h.getFile(w, r) }

func (h *handlers) pushChunk(w http.ResponseWriter, r *http.Request) {
	var (
		q           = r.URL.Query()
		taskID, err = uuid.Parse(q.Get("task_id"))
		quality     = q.Get("quality")
		chunkNum    int
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if chunkNum, err = strconv.Atoi(q.Get("chunk_num")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if quality == "" || chunkNum < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l := log.WithFields(log.Fields{
		"mod":     "http",
		"task_id": taskID,
		"chunk":   chunkNum,
		"quality": quality,
	})

	var dstPath = filepath.Join(
		h.workDir,
		taskID.String(),
		"chunks",
		quality,
		fmt.Sprintf("chunk_%03d.mp4", chunkNum),
	)
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := os.Create(dstPath)
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	removeDst := false
	defer func() {
		f.Sync()
		f.Close()
		if removeDst {
			if err := os.Remove(dstPath); err != nil {
				l.Errorf("remove source due to error: %v", err)
			}
		}
	}()

	n, err := io.Copy(f, r.Body)
	if err != nil {
		removeDst = true
		l.Errorf("copy file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n != r.ContentLength {
		removeDst = true
		l.Errorf("content length header didn't match file size: %v vs %v", r.ContentLength, n)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handlers) pushPoster(w http.ResponseWriter, r *http.Request) {
	var (
		q           = r.URL.Query()
		taskID, err = uuid.Parse(q.Get("task_id"))
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l := log.WithFields(log.Fields{
		"mod":     "http",
		"task_id": taskID,
	})

	var dstPath = filepath.Join(
		h.workDir,
		taskID.String(),
		"poster",
		"poster.jpg",
	)
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := os.Create(dstPath)
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	removeDst := false
	defer func() {
		f.Sync()
		f.Close()
		if removeDst {
			if err := os.Remove(dstPath); err != nil {
				l.Errorf("remove source due to error: %v", err)
			}
		}
	}()

	n, err := io.Copy(f, r.Body)
	if err != nil {
		removeDst = true
		l.Errorf("copy file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n != r.ContentLength {
		removeDst = true
		l.Errorf("content length header didn't match file size: %v vs %v", r.ContentLength, n)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handlers) getAudio(w http.ResponseWriter, r *http.Request) { h.getFile(w, r) }

func (h *handlers) pushAudio(w http.ResponseWriter, r *http.Request) {
	var (
		q           = r.URL.Query()
		taskID, err = uuid.Parse(q.Get("task_id"))
		trackNum    int
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if trackNum, err = strconv.Atoi(q.Get("track_num")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if trackNum < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l := log.WithFields(log.Fields{
		"mod":     "http",
		"task_id": taskID,
		"chunk":   trackNum,
		"quality": "audio",
	})

	var dstPath = filepath.Join(
		h.workDir,
		taskID.String(),
		"audios",
		fmt.Sprintf("audio_%d.mp4", trackNum),
	)
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := os.Create(dstPath)
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	removeDst := false
	defer func() {
		f.Sync()
		f.Close()
		if removeDst {
			if err := os.Remove(dstPath); err != nil {
				l.Errorf("remove source due to error: %v", err)
			}
		}
	}()

	n, err := io.Copy(f, r.Body)
	if err != nil {
		removeDst = true
		l.Errorf("copy file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if n != r.ContentLength {
		removeDst = true
		l.Errorf("content length header didn't match file size: %v vs %v", r.ContentLength, n)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
