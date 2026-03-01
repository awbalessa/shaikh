package sse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Writer struct {
	w *bufio.Writer
	fl http.Flusher
}

func New(w http.ResponseWriter) (*Writer, error) {
	fl, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("response writer does not support flushing")
	}

	return &Writer{
		w: bufio.NewWriter(w),
		fl: fl,
	}, nil
}

func (s *Writer) SendJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", b); err != nil {
		return err
	}

	if err := s.w.Flush(); err != nil {
		return err
	}

	s.fl.Flush()
	return nil
}

func (s *Writer) Done() error {
	if _, err := fmt.Fprint(s.w, "data: [DONE]\n\n"); err != nil {
		return err
	}

	if err := s.w.Flush(); err != nil {
		return err
	}

	s.fl.Flush()
	return nil
}
