package opennotify

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes opennotify as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/opennotify-cli/opennotify"
//
// The same Domain also builds the standalone opennotify binary (see main.go).
func init() { kit.Register(Domain{}) }

// Domain is the opennotify driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "opennotify",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "opennotify",
			Short:  "ISS location and people in space (Open Notify)",
			Long: `opennotify fetches real-time ISS position and the current list of
humans in space from the Open Notify API. No API key required.`,
			Site: Host,
			Repo: "https://github.com/tamnd/opennotify-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// position: current ISS position (single record)
	kit.Handle(app, kit.OpMeta{
		Name:    "position",
		Group:   "read",
		Single:  true,
		Summary: "Show the current ISS position",
	}, positionOp)

	// astronauts: people currently in space
	kit.Handle(app, kit.OpMeta{
		Name:    "astronauts",
		Group:   "read",
		List:    true,
		Summary: "List humans currently in space",
	}, astronautsOp)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type positionInput struct {
	Client *Client `kit:"inject"`
}

type astronautsInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results (0 = all)"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func positionOp(ctx context.Context, in positionInput, emit func(*Position) error) error {
	pos, err := in.Client.Position(ctx)
	if err != nil {
		return err
	}
	return emit(pos)
}

func astronautsOp(ctx context.Context, in astronautsInput, emit func(Astronaut) error) error {
	people, err := in.Client.Astronauts(ctx)
	if err != nil {
		return err
	}
	n := 0
	for _, p := range people {
		if in.Limit > 0 && n >= in.Limit {
			break
		}
		if err := emit(p); err != nil {
			return err
		}
		n++
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty opennotify reference")
	}
	return "position", input, nil
}

// Locate returns the live URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "position":
		return "http://" + Host + "/iss-now.json", nil
	case "astronauts":
		return "http://" + Host + "/astros.json", nil
	default:
		return "", errs.Usage("opennotify has no resource type %q", uriType)
	}
}
