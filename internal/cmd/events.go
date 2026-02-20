package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// EventsCmd is the parent command for websocket event streaming.
type EventsCmd struct {
	Tail EventsTailCmd `cmd:"" help:"Follow live websocket events"`
}

// EventsTailCmd streams live events from GET /v1/ws.
type EventsTailCmd struct {
	ChatIDs        []string      `help:"Subscribe to specific chat IDs (repeatable)" name:"chat-id"`
	All            bool          `help:"Subscribe to all chats" name:"all"`
	IncludeControl bool          `help:"Include control messages (ready, subscriptions.updated, error)" name:"include-control"`
	Reconnect      bool          `help:"Reconnect on disconnect/errors" default:"true"`
	ReconnectDelay time.Duration `help:"Delay before reconnect attempts" name:"reconnect-delay" default:"2s"`
	StopAfter      time.Duration `help:"Stop after duration (0=forever)" name:"stop-after" default:"0s"`
}

// Run executes the events tail command.
func (c *EventsTailCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.StopAfter < 0 {
		return errfmt.UsageError("invalid --stop-after %s (must be >= 0)", c.StopAfter)
	}
	if c.Reconnect && c.ReconnectDelay <= 0 {
		return errfmt.UsageError("invalid --reconnect-delay %s (must be > 0)", c.ReconnectDelay)
	}
	chatIDs, err := resolveEventSubscriptions(c.All, c.ChatIDs)
	if err != nil {
		return err
	}

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	stopDeadline := time.Time{}
	if c.StopAfter > 0 {
		stopDeadline = time.Now().Add(c.StopAfter)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	for {
		if stopReached(stopDeadline) {
			return nil
		}

		conn, err := client.Events().Connect(ctx)
		if err != nil {
			if beeperapi.IsEventsUnsupported(err) {
				return errUnsupportedWebsocketEvents()
			}
			if !c.Reconnect {
				return err
			}
			if err := waitForReconnect(ctx, c.ReconnectDelay, stopDeadline); err != nil {
				return nil
			}
			continue
		}

		if err := conn.SetSubscriptions(ctx, "", chatIDs); err != nil {
			_ = conn.Close()
			return err
		}

		if err := c.readLoop(ctx, stopDeadline, conn, encoder, u); err != nil {
			_ = conn.Close()
			if isEventsStreamClosed(err) {
				if !c.Reconnect {
					return nil
				}
				if err := waitForReconnect(ctx, c.ReconnectDelay, stopDeadline); err != nil {
					return nil
				}
				continue
			}
			if !c.Reconnect {
				return err
			}
			if err := waitForReconnect(ctx, c.ReconnectDelay, stopDeadline); err != nil {
				return nil
			}
			continue
		}

		_ = conn.Close()
		return nil
	}
}

func (c *EventsTailCmd) readLoop(ctx context.Context, stopDeadline time.Time, conn *beeperapi.EventsConnection, encoder *json.Encoder, u *ui.UI) error {
	for {
		if stopReached(stopDeadline) {
			return nil
		}

		readCtx, cancel := nextEventsReadContext(ctx, stopDeadline)
		evt, err := conn.ReadEvent(readCtx)
		deadlineReached := errors.Is(readCtx.Err(), context.DeadlineExceeded)
		cancel()
		if err != nil {
			if isReadTimeout(err) {
				// WebSocket reads are not safely repeatable after a read deadline timeout.
				// Only treat timeout as graceful completion when we've reached --stop-after.
				if stopReached(stopDeadline) || (deadlineReached && !stopDeadline.IsZero()) {
					return nil
				}
			}
			return err
		}

		if !c.IncludeControl && evt.IsControlMessage() {
			continue
		}
		if err := writeEventOutput(ctx, encoder, u, evt); err != nil {
			return err
		}
	}
}

func resolveEventSubscriptions(all bool, chatIDs []string) ([]string, error) {
	if all && len(chatIDs) > 0 {
		return nil, errfmt.UsageError("cannot combine --all with --chat-id")
	}

	if all {
		return []string{"*"}, nil
	}

	if len(chatIDs) == 0 {
		return nil, errfmt.UsageError("select subscriptions with --all or --chat-id")
	}

	trimmed := make([]string, 0, len(chatIDs))
	for _, chatID := range chatIDs {
		value := strings.TrimSpace(chatID)
		if value == "" {
			return nil, errfmt.UsageError("--chat-id cannot be empty")
		}
		trimmed = append(trimmed, value)
	}

	hasWildcard := false
	for _, chatID := range trimmed {
		if chatID == "*" {
			hasWildcard = true
			break
		}
	}
	if hasWildcard && len(trimmed) > 1 {
		return nil, errfmt.UsageError(`cannot combine "*" with specific --chat-id values`)
	}

	return trimmed, nil
}

func writeEventOutput(ctx context.Context, encoder *json.Encoder, u *ui.UI, evt beeperapi.Event) error {
	switch {
	case outfmt.IsJSON(ctx):
		if err := encoder.Encode(evt); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
	case outfmt.IsPlain(ctx):
		u.Out().Printf("%s\t%d\t%d\t%s\t%s", evt.Type, evt.Seq, evt.TS, evt.ChatID, strings.Join(evt.IDs, ","))
	default:
		if evt.IsControlMessage() {
			u.Out().Dim(fmt.Sprintf("[control] %s %s", evt.Type, evt.Message))
			return nil
		}
		u.Out().Printf("[%s] chat=%s seq=%d ids=%s", evt.Type, evt.ChatID, evt.Seq, strings.Join(evt.IDs, ","))
	}
	return nil
}

func waitForReconnect(ctx context.Context, delay time.Duration, stopDeadline time.Time) error {
	if stopReached(stopDeadline) {
		return context.DeadlineExceeded
	}
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < delay {
			delay = remaining
		}
	}
	if !stopDeadline.IsZero() {
		remaining := time.Until(stopDeadline)
		if remaining < delay {
			delay = remaining
		}
	}
	if delay <= 0 {
		return context.DeadlineExceeded
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func nextEventsReadContext(ctx context.Context, stopDeadline time.Time) (context.Context, context.CancelFunc) {
	ctxDeadline, hasCtxDeadline := ctx.Deadline()
	switch {
	case !stopDeadline.IsZero() && (!hasCtxDeadline || stopDeadline.Before(ctxDeadline)):
		return context.WithDeadline(ctx, stopDeadline)
	case hasCtxDeadline:
		return context.WithDeadline(ctx, ctxDeadline)
	default:
		return ctx, func() {}
	}
}

func stopReached(deadline time.Time) bool {
	return !deadline.IsZero() && time.Now().After(deadline)
}

func isReadTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isEventsStreamClosed(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return true
	}
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
	) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "unexpected eof")
}

func errUnsupportedWebsocketEvents() error {
	return fmt.Errorf("websocket events are not supported by this Beeper Desktop API version (requires a newer Beeper Desktop build)")
}
