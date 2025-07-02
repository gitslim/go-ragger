package rag

import (
	"context"
	"log/slog"
	"time"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type RAGPipeline struct {
	pool   *pgxpool.Pool
	db     *sqlc.Queries
	logger *slog.Logger
	cfg    *RAGConfig
}

type RAGConfig struct {
	BatchSize     int
	PollInterval  time.Duration
	StuckTimeout  time.Duration
	WorkerCount   int
	ProcessingTTL time.Duration
}

func NewRAGPipeline(pool *pgxpool.Pool, db *sqlc.Queries, logger *slog.Logger, cfg *RAGConfig) *RAGPipeline {
	return &RAGPipeline{
		pool:   pool,
		db:     db,
		logger: logger.With("component", "rag_pipeline"),
		cfg:    cfg,
	}
}

func (p *RAGPipeline) Run(ctx context.Context) error {
	// Запускаем воркеры
	for i := range p.cfg.WorkerCount {
		go p.worker(ctx, i)
	}

	<-ctx.Done()
	return nil
}

func (p *RAGPipeline) worker(ctx context.Context, workerID int) {
	logger := p.logger.With("worker_id", workerID)
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.documentIndexerBatch(ctx); err != nil {
				logger.Error("failed to process sender batch", "error", err)
			}
		}
	}
}

func RunRAGPipeline(lc fx.Lifecycle, logger *slog.Logger, pool *pgxpool.Pool, db *sqlc.Queries) {
	cfg := &RAGConfig{
		BatchSize:     10,
		PollInterval:  5 * time.Second,
		StuckTimeout:  60 * time.Minute,
		WorkerCount:   3,
		ProcessingTTL: 30 * time.Minute,
	}

	processor := NewRAGPipeline(pool, db, logger, cfg)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := processor.Run(context.Background()); err != nil {
					logger.Error("rag pipeline failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
