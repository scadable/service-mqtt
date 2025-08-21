package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	server "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	packets "github.com/mochi-mqtt/server/v2/packets"
	"github.com/rs/zerolog"

	"service-mqtt/internal/config"
	"service-mqtt/internal/core/devices"
	"service-mqtt/internal/core/port"
)

// NewServer creates and configures a new MQTT server instance.
func NewServer(ctx context.Context, cfg config.Config, log zerolog.Logger, pub port.Publisher, mgr *devices.Manager) (*server.Server, error) {
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	s := server.New(&server.Options{Logger: slogLogger})

	ah := &AuthHook{
		mgr: mgr,
		log: log.With().Str("hook", "auth").Logger(),
	}
	if err := s.AddHook(ah, nil); err != nil {
		return nil, err
	}

	mh := &MessageHook{}
	opts := &messageHookOptions{
		ctx:       ctx,
		log:       log.With().Str("hook", "message").Logger(),
		publisher: pub,
	}
	if err := s.AddHook(mh, opts); err != nil {
		return nil, err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: fmt.Sprintf(":%s", cfg.MQTTPort),
	})
	if err := s.AddListener(tcp); err != nil {
		return nil, err
	}

	return s, nil
}

/* ------------------------ AUTH HOOK --------------------------------------- */
type AuthHook struct {
	server.HookBase
	mgr *devices.Manager
	log zerolog.Logger
}

// Provides now includes OnACLCheck to handle authorization.
func (h *AuthHook) Provides(b byte) bool {
	switch b {
	case server.OnConnectAuthenticate, server.OnACLCheck:
		return true
	default:
		return false
	}
}

func (h *AuthHook) OnConnectAuthenticate(cl *server.Client, pk packets.Packet) bool {
	if h.mgr.Authenticate(string(pk.Connect.Username), string(pk.Connect.Password)) {
		h.log.Info().Str("client_id", cl.ID).Msg("client authenticated")
		return true
	}

	h.log.Warn().Str("client_id", cl.ID).Msg("bad credentials")
	return false
}

// OnACLCheck allows any authenticated client to publish and subscribe.
func (h *AuthHook) OnACLCheck(cl *server.Client, topic string, write bool) bool {
	return true
}

/* ---------------------- MESSAGE HOOK -------------------------------------- */
type messageHookOptions struct {
	ctx       context.Context
	log       zerolog.Logger
	publisher port.Publisher
}

type MessageHook struct {
	server.HookBase
	opts *messageHookOptions
}

func (h *MessageHook) Init(cfg any) error {
	if o, ok := cfg.(*messageHookOptions); ok {
		h.opts = o
	}
	return nil
}

// Provides now includes OnPublish to handle incoming messages.
func (h *MessageHook) Provides(b byte) bool {
	return b == server.OnPublish
}

// OnPublish forwards the raw payload to NATS.
func (h *MessageHook) OnPublish(c *server.Client, pk packets.Packet) (packets.Packet, error) {
	// The username is attached to the client's properties after authentication.
	subject := fmt.Sprintf("raw.%s", c.Properties.Username)
	if err := h.opts.publisher.Publish(h.opts.ctx, subject, pk.Payload); err != nil {
		h.opts.log.Error().Err(err).Msg("NATS publish error")
		return pk, err
	}

	h.opts.log.Info().
		Int("bytes", len(pk.Payload)).
		Str("subject", subject).
		Str("payload", string(pk.Payload)).
		Msg("forwarded to NATS")
	return pk, nil
}
