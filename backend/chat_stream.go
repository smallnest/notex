package backend

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kataras/golog"
)

var chatStreamUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleChatStream 使用 WebSocket 输出对话流式结果。
func (s *Server) handleChatStream(c *gin.Context) {
	notebookID := c.Param("id")
	conn, err := chatStreamUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		golog.Errorf("failed to upgrade websocket: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})

	for {
		var req ChatRequest
		if err := conn.ReadJSON(&req); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				golog.Errorf("chat stream read error: %v", err)
			}
			return
		}

		req.Message = strings.TrimSpace(req.Message)
		if req.Message == "" {
			_ = conn.WriteJSON(ChatStreamEvent{Type: "error", Error: "message required"})
			continue
		}

		if err := s.loadNotebookVectorIndex(ctx, notebookID); err != nil {
			golog.Errorf("failed to load vector index: %v", err)
		}

		sessionID := req.SessionID
		if sessionID == "" {
			session, err := s.store.CreateChatSession(ctx, notebookID, "")
			if err != nil {
				_ = conn.WriteJSON(ChatStreamEvent{Type: "error", Error: "Failed to create session"})
				continue
			}
			sessionID = session.ID
		}

		if _, err := s.store.AddChatMessage(ctx, sessionID, "user", req.Message, nil); err != nil {
			_ = conn.WriteJSON(ChatStreamEvent{Type: "error", Error: "Failed to add message"})
			continue
		}

		session, err := s.store.GetChatSession(ctx, sessionID)
		if err != nil {
			_ = conn.WriteJSON(ChatStreamEvent{Type: "error", Error: "Failed to get session"})
			continue
		}

		response, err := s.agent.ChatStream(ctx, notebookID, req.Message, session.Messages, func(delta string) error {
			if delta == "" {
				return nil
			}
			return conn.WriteJSON(ChatStreamEvent{Type: "delta", Content: delta})
		})
		if err != nil {
			_ = conn.WriteJSON(ChatStreamEvent{Type: "error", Error: err.Error()})
			continue
		}

		sourceIDs := make([]string, len(response.Sources))
		for i, src := range response.Sources {
			sourceIDs[i] = src.ID
		}

		saved, saveErr := s.store.AddChatMessage(ctx, sessionID, "assistant", response.Message, sourceIDs)
		if saveErr != nil {
			golog.Errorf("failed to save chat message: %v", saveErr)
		}

		response.SessionID = sessionID
		if saved != nil {
			response.MessageID = saved.ID
		}

		doneEvent := ChatStreamEvent{
			Type:      "done",
			Message:   response.Message,
			Sources:   response.Sources,
			SessionID: response.SessionID,
			MessageID: response.MessageID,
		}

		if err := conn.WriteJSON(doneEvent); err != nil {
			golog.Errorf("failed to write chat done event: %v", err)
			return
		}
	}
}
