package chunker

import (
	"context"
	"log/slog"
	"time"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

// Chunker is a struct that contains components to process chunks of data.
type Chunker struct {
	pool           *pgxpool.Pool
	db             *sqlc.Queries
	logger         *slog.Logger
	cfg            *ChunkerConfig
	chunkrAIClient *ChunkrAIClient
}

// ChunkerConfig is a struct that contains configuration for the chunker.
type ChunkerConfig struct {
	BatchSize     int
	PollInterval  time.Duration
	StuckTimeout  time.Duration
	WorkerCount   int
	ProcessingTTL time.Duration
}

// NewChunker creates a new instance of Chunker.
func NewChunker(pool *pgxpool.Pool, db *sqlc.Queries, logger *slog.Logger, cfg *ChunkerConfig, client *ChunkrAIClient) *Chunker {
	return &Chunker{
		pool:           pool,
		db:             db,
		logger:         logger.With("component", "chunker"),
		cfg:            cfg,
		chunkrAIClient: client,
	}
}

// Run starts chunking workers and a stuck documents cleaner.
func (p *Chunker) Run(ctx context.Context) error {
	// Запускаем воркеры
	for i := range p.cfg.WorkerCount {
		go p.worker(ctx, i)
	}

	// Запускаем очистку зависших задач
	go p.stuckDocumentsCleaner(ctx)

	<-ctx.Done()
	return nil
}

// worker periodically sends and checks batches of documents.
func (p *Chunker) worker(ctx context.Context, workerID int) {
	logger := p.logger.With("worker_id", workerID)
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.senderBatch(ctx, logger); err != nil {
				logger.Error("sender batch failed", "error", err)
			}
			if err := p.checkerBatch(ctx, logger); err != nil {
				logger.Error("checker batch failed", "error", err)
			}
		}
	}
}

// RunChunker creates a new chunker and runs it
func RunChunker(lc fx.Lifecycle, logger *slog.Logger, pool *pgxpool.Pool, db *sqlc.Queries, client *ChunkrAIClient) {
	cfg := &ChunkerConfig{
		BatchSize:     10,
		PollInterval:  5 * time.Second,
		StuckTimeout:  60 * time.Minute,
		WorkerCount:   3,
		ProcessingTTL: 30 * time.Minute,
	}

	chunker := NewChunker(pool, db, logger, cfg, client)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := chunker.Run(context.Background()); err != nil {
					logger.Error("chunker run failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
