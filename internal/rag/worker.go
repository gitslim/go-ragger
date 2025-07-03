package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type RAGPipeline struct {
	pool    *pgxpool.Pool
	db      *sqlc.Queries
	logger  *slog.Logger
	cfg     *RAGConfig
	indexer indexer.Indexer
}

type RAGConfig struct {
	BatchSize    int
	PollInterval time.Duration
	WorkerCount  int
}

func NewRAGPipeline(pool *pgxpool.Pool, db *sqlc.Queries, logger *slog.Logger, cfg *RAGConfig, indexer indexer.Indexer) *RAGPipeline {
	return &RAGPipeline{
		pool:    pool,
		db:      db,
		logger:  logger.With("component", "rag_pipeline"),
		cfg:     cfg,
		indexer: indexer,
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
			if err := p.indexerBatch(ctx); err != nil {
				logger.Error("failed to process indexer batch", "error", err)
			}
		}
	}
}

func (p *RAGPipeline) indexerBatch(ctx context.Context) error {
	docs, err := p.db.GetChunkedDocuments(ctx, int32(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("get chunked documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.indexDocument(ctx, doc.ID); err != nil {
			p.logger.Error("failed to index document",
				"doc_id", doc.ID,
				"error", err)

			// if err := p.db.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
			// 	ID:     doc.ID,
			// 	Status: sqlc.DocumentStatusIndexfail,
			// }); err != nil {
			// 	p.logger.Error("failed to mark document as indexfail",
			// 		"doc_id", doc.ID,
			// 		"error", err)
			// }
		}
	}

	return nil
}

func (p *RAGPipeline) indexDocument(ctx context.Context, docID uuid.UUID) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin indexing transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := p.db.WithTx(tx)

	doc, err := q.LockDocumentForIndexing(ctx, docID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("lock document for indexing: %w", err)
	}

	p.logger.Info("indexing document",
		"doc_id", doc.ID,
		"file_name", doc.FileName)

	err = p.storeDocumentToIndex(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to store document to index: %w", err)
	}
	p.logger.Info("document indexed",
		"doc_id", doc.ID,
		"file_name", doc.FileName)

	if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
		ID:     doc.ID,
		Status: sqlc.DocumentStatusIndexed,
	}); err != nil {
		p.logger.Error("failed to mark document as indexed",
			"doc_id", doc.ID,
			"error", err)
	}

	return tx.Commit(ctx)
}

func (p *RAGPipeline) storeDocumentToIndex(ctx context.Context, doc sqlc.Document) error {

	var data *chunkrai.TaskResponse

	err := json.Unmarshal(doc.ChunkrResult, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshall chunkr data: %w", err)
	}

	chunks := make([]*schema.Document, 0)
	for _, chunk := range data.Output.Chunks {
		chunkStrings := make([]string, 0)

		for _, segment := range chunk.Segments {
			chunkStrings = append(chunkStrings, *segment.Markdown)
		}
		content := strings.Join(chunkStrings, "\n")

		chunks = append(chunks, &schema.Document{
			ID:      *chunk.ChunkId,
			Content: content,
			MetaData: map[string]any{
				"doc_id":       doc.ID.String(),
				"doc_filename": doc.FileName,
			},
		})
	}

	_, err = p.indexer.Store(ctx, chunks)
	if err != nil {
		return fmt.Errorf("indexer store error: %w", err)
	}

	return nil
}

func RunRAGPipeline(lc fx.Lifecycle, logger *slog.Logger, pool *pgxpool.Pool, db *sqlc.Queries, indexer indexer.Indexer) {
	cfg := &RAGConfig{
		BatchSize:    10,
		PollInterval: 5 * time.Second,
		WorkerCount:  3,
	}

	pipeline := NewRAGPipeline(pool, db, logger, cfg, indexer)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := pipeline.Run(context.Background()); err != nil {
					logger.Error("rag pipeline run failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
