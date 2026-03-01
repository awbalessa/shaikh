package chat

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/awbalessa/shaikh/api/internal/app/ai"
	"github.com/awbalessa/shaikh/api/internal/http/sse"
)

type Handler struct {
	model ai.LModel
}

func New(model ai.LModel) *Handler {
	return &Handler{
		model: model,
	}
}

type ChatRequest struct {
	Messages any `json:"messages"`
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("x-vercel-ai-ui-message-stream", "v1")

	enc, err := sse.New(w)
	if err != nil {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = enc.SendJSON(map[string]any{
			"type": "error",
			"errorText": "bad request",
		})
		_ = enc.Done()
		return
	}

	stream, err := h.model.Stream(r.Context(), nil)
	if err != nil {
		_ = enc.SendJSON(map[string]any{
			"type": "error",
			"errorText": safeErrText(err),
		})
		_ = enc.Done()
		return
	}
	defer stream.Close()

	for {
		ev, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				_ = enc.Done()
				return
			}

			_ = enc.SendJSON(map[string]any{
				"type": "error",
				"errorText": safeErrText(err),
			})
			_ = enc.Done()
			return
		}

		part, ok := toVercelPart(ev)
		if ok {
			if err := enc.SendJSON(part); err != nil {
				return
			}
		}

		if ev.Type = ai.EventFinish {
			_ = enc.Done()
			return
		}
	}
}

func toVercelPart(ev ai.Event) (map[string]any, bool) {
	switch ev.Type {
		case ai.EventStreamStart:
		return map[string]any{
			"type": "start",
			"messageId": "msg_server_1",
		}, true
		case ai.EventTextStart:
		return map[string]any{
			"type": "text-start",
			"id": ev.ID,
		}, true
		case ai.EventTextDelta:
		return map[string]any{
			"type": "text-delta",
			"id": ev.ID,
			"delta": ev.Delta,
		}, true
		case ai.EventTextEnd:
		return map[string]any{
			"type": "text-end",
			"id": ev.ID,
		}, true
		case ai.EventFinish:
		return map[string]any{
			"type": "finish",
		}, true
		case ai.EventError:
		return map[string]any{
			"type": "error",
			"errorText": safeErrText(ev.Err),
		}, true
	}
	return nil, false
}

func safeErrText(err error) string {
	if err == nil {
		return "unknown error"
	}
	return err.Error()
}
