// Package orders is the composition root for the orders module.
package orders

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	idp "avex-backend/internal/modules/identity/port"
	"avex-backend/internal/modules/orders/events"
	"avex-backend/internal/modules/orders/port"
	"avex-backend/internal/modules/orders/repository/postgres"
	"avex-backend/internal/modules/orders/service"
	httptransport "avex-backend/internal/modules/orders/transport/http"
	"avex-backend/internal/platform/config"
)

// Module is the wired orders module.
type Module struct {
	svc       port.ServicePort
	pool      *pgxpool.Pool
	logger    *slog.Logger
	jwtIssuer idp.JWTIssuer
}

// New wires all orders dependencies and returns a ready-to-use Module.
func New(cfg *config.Config, pool *pgxpool.Pool, logger *slog.Logger, jwtIssuer idp.JWTIssuer) *Module {
	// Repositories
	repos := postgres.NewRepositories()
	repoSet := repos.RepositorySet()

	// Event Publisher
	eventPublisher := events.NewEventPublisher(repoSet, &uuidIDGen{})

	// TxRunner
	txRunner := &pgxTxRunner{pool: pool}

	// Service
	deps := port.Deps{
		Clock:                &realClock{},
		IDGenerator:          &uuidIDGen{},
		OrderNumberGenerator: &orderNumberGen{},
		EventPublisher:       eventPublisher,
		Logger:               logger,
		TxRunner:             txRunner,
		Repos:                repoSet,
	}

	svc := service.New(deps, pool, service.Config{
		OfferTTL: 15 * time.Second,
	})

	return &Module{svc: svc, pool: pool, logger: logger, jwtIssuer: jwtIssuer}
}

func (m *Module) Service() port.ServicePort { return m.svc }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	httptransport.RegisterRoutes(mux, m.svc, m.logger, m.jwtIssuer)
}

func (m *Module) Close() {}

// ===== Adapters =====

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type uuidIDGen struct{}

func (*uuidIDGen) NewID() string { return newUUID() }

type orderNumberGen struct{}

func (*orderNumberGen) Generate() string {
	return "AVEX-" + time.Now().UTC().Format("20060102") + "-" + newUUID()[:8]
}

type pgxTxRunner struct{ pool *pgxpool.Pool }

func (r *pgxTxRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, exec port.Executor) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
