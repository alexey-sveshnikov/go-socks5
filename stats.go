package socks5

import (
	"log"
	"time"
)

type EventsHandler interface {
	OnSessionStarted(request *Request)
	OnSessionFinished(request *Request, sessionLength time.Duration)
	OnSessionBlocked(request *Request)
	OnProxiedConnectionStarted(request *Request, conn conn)
	OnUploadBytes(request *Request, bytes int64)
	OnDownloadBytes(request *Request, bytes int64)
}

// Basic EventHandler just logs everything
type LoggingEventsHandler struct {
	logger *log.Logger
}

func (h LoggingEventsHandler) getLoggingPrefix(request *Request) string {
	prefix := "[STAT] User"
	if request.AuthContext != nil {
		prefix += " " + request.AuthContext.Username()
	}
	return prefix
}

func (h LoggingEventsHandler) OnSessionStarted(request *Request) {
	prefix := h.getLoggingPrefix(request)
	h.logger.Printf("%s connected\n", prefix)
}

func (h LoggingEventsHandler) OnSessionFinished(request *Request, sessionLength time.Duration) {
	prefix := h.getLoggingPrefix(request)
	h.logger.Printf("%s disconnected, session length: %.2f secs\n", prefix, sessionLength.Seconds())
}

func (h LoggingEventsHandler) OnSessionBlocked(request *Request) {
	// Since this is an error condition in Server.handleConnect,
	// it's already being logging there
}

func (h LoggingEventsHandler) OnUploadBytes(request *Request, bytes int64) {
	prefix := h.getLoggingPrefix(request)
	h.logger.Printf("%s uploaded %v bytes\n", prefix, bytes)
}

func (h LoggingEventsHandler) OnDownloadBytes(request *Request, bytes int64) {
	prefix := h.getLoggingPrefix(request)
	h.logger.Printf("%s downloaded %v bytes\n", prefix, bytes)
}

func (h LoggingEventsHandler) OnProxiedConnectionStarted(request *Request, conn conn) {
	prefix := h.getLoggingPrefix(request)
	h.logger.Printf("%s connect %s -> %s:%d\n",
		prefix, conn.RemoteAddr(),
		request.realDestAddr.IP, request.realDestAddr.Port)
}

// Dispatcher that send events to multiple handlers
type EventDispatcher struct {
	handlers [] EventsHandler
}

func (d EventDispatcher) OnSessionStarted(request *Request) {
	for _, handler := range d.handlers {
		handler.OnSessionStarted(request)
	}
}

func (d EventDispatcher) OnSessionFinished(request *Request, sessionLength time.Duration) {
	for _, handler := range d.handlers {
		handler.OnSessionFinished(request, sessionLength)
	}
}

func (d EventDispatcher) OnSessionBlocked(request *Request) {
	for _, handler := range d.handlers {
		handler.OnSessionBlocked(request)
	}
}

func (d EventDispatcher) OnUploadBytes(request *Request, bytes int64) {
	for _, handler := range d.handlers {
		handler.OnUploadBytes(request, bytes)
	}
}

func (d EventDispatcher) OnDownloadBytes(request *Request, bytes int64) {
	for _, handler := range d.handlers {
		handler.OnDownloadBytes(request, bytes)
	}
}

func (d EventDispatcher) OnProxiedConnectionStarted(request *Request, conn conn) {
	for _, handler := range d.handlers {
		handler.OnProxiedConnectionStarted(request, conn)
	}
}
