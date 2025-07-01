package chunkr

import (
	"context"
	"log/slog"
	"time"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type DocumentProcessor struct {
	pool   *pgxpool.Pool
	db     *sqlc.Queries
	logger *slog.Logger
	cfg    *ProcessorConfig
	client *Client
}

type ProcessorConfig struct {
	BatchSize     int
	PollInterval  time.Duration
	StuckTimeout  time.Duration
	WorkerCount   int
	ProcessingTTL time.Duration
}

func NewDocumentProcessor(pool *pgxpool.Pool, db *sqlc.Queries, logger *slog.Logger, cfg *ProcessorConfig, client *Client) *DocumentProcessor {
	return &DocumentProcessor{
		pool:   pool,
		db:     db,
		logger: logger.With("component", "document_processor"),
		cfg:    cfg,
		client: client,
	}
}

func (p *DocumentProcessor) Run(ctx context.Context) error {
	// Запускаем воркеры
	for i := range p.cfg.WorkerCount {
		go p.worker(ctx, i)
	}

	// Запускаем очистку зависших задач
	go p.stuckDocumentsCleaner(ctx)

	<-ctx.Done()
	return nil
}

func (p *DocumentProcessor) worker(ctx context.Context, workerID int) {
	logger := p.logger.With("worker_id", workerID)
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.documentSenderBatch(ctx, logger); err != nil {
				logger.Error("failed to process sender batch", "error", err)
			}
			if err := p.documentCheckerBatch(ctx, logger); err != nil {
				logger.Error("failed to process checker batch", "error", err)
			}
		}
	}
}

func RunDocumentProcessor(lc fx.Lifecycle, logger *slog.Logger, pool *pgxpool.Pool, db *sqlc.Queries, client *Client) {
	cfg := &ProcessorConfig{
		BatchSize:     10,
		PollInterval:  5 * time.Second,
		StuckTimeout:  60 * time.Minute,
		WorkerCount:   3,
		ProcessingTTL: 30 * time.Minute,
	}

	processor := NewDocumentProcessor(pool, db, logger, cfg, client)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := processor.Run(context.Background()); err != nil {
					logger.Error("document processor failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
