package indexator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/vectordb/milvus"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

// Indexator contains components for indexing documents into a vector database.
type Indexator struct {
	db             *pgxpool.Pool
	q              *sqlc.Queries
	logger         *slog.Logger
	config         *IndexatorConfig
	indexerFactory milvus.MilvusIndexerFactory
}

// IndexatorConfig is the configuration for the indexator.
type IndexatorConfig struct {
	BatchSize    int
	PollInterval time.Duration
	WorkerCount  int
}

// NewIndexator creates a new indexator.
func NewIndexator(db *pgxpool.Pool, q *sqlc.Queries, logger *slog.Logger, cfg *IndexatorConfig, indexerFactory milvus.MilvusIndexerFactory) *Indexator {
	return &Indexator{
		db:             db,
		q:              q,
		logger:         logger.With("component", "indexator"),
		config:         cfg,
		indexerFactory: indexerFactory,
	}
}

// Run starts the indexator.
func (p *Indexator) Run(ctx context.Context) error {
	// run workers
	for i := range p.config.WorkerCount {
		go p.worker(ctx, i)
	}

	<-ctx.Done()
	return nil
}

// worker periodically runs indexer batches
func (p *Indexator) worker(ctx context.Context, workerID int) {
	logger := p.logger.With("worker_id", workerID)
	ticker := time.NewTicker(p.config.PollInterval)
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

// indexerBatch fetches documents from the queue and indexes them
func (p *Indexator) indexerBatch(ctx context.Context) error {
	docs, err := p.q.GetChunkedDocuments(ctx, int32(p.config.BatchSize))
	if err != nil {
		return fmt.Errorf("get chunked documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.indexDocument(ctx, doc.ID); err != nil {
			p.logger.Error("failed to index document",
				"doc_id", doc.ID,
				"error", err)
		}
	}

	return nil
}

// indexDocument transactionally indexes a document and update status
func (p *Indexator) indexDocument(ctx context.Context, docID uuid.UUID) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin indexing transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := p.q.WithTx(tx)

	doc, err := q.LockDocumentForIndexing(ctx, docID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("lock document for indexing: %w", err)
	}

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

// storeDocumentToIndex stores document to indexer store
func (p *Indexator) storeDocumentToIndex(ctx context.Context, doc sqlc.Document) error {

	var chunkrResponse *chunkrai.TaskResponse

	err := json.Unmarshal(doc.ChunkrResult, &chunkrResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshall chunkr data: %w", err)
	}

	chunks := make([]*schema.Document, 0)
	for _, chunk := range chunkrResponse.Output.Chunks {
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

	indexer, err := p.indexerFactory(ctx, &milvus.MilvusIndexerConfig{Collection: milvus.ToMilvusName(doc.UserID.String())})
	if err != nil {
		return fmt.Errorf("failed to create indexer: %w", err)
	}

	p.logger.Info("indexing document", "doc_id", doc.ID, "file_name", doc.FileName, "chunks", len(chunks))

	_, err = indexer.Store(ctx, chunks)
	if err != nil {
		return fmt.Errorf("indexer store error: %w", err)
	}

	return nil
}

// RunIndexator creates and starts the indexator
func RunIndexator(lc fx.Lifecycle, logger *slog.Logger, pool *pgxpool.Pool, db *sqlc.Queries, indexerFactory milvus.MilvusIndexerFactory) {
	cfg := &IndexatorConfig{
		BatchSize:    10,
		PollInterval: 5 * time.Second,
		WorkerCount:  3,
	}

	indexator := NewIndexator(pool, db, logger, cfg, indexerFactory)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := indexator.Run(context.Background()); err != nil {
					logger.Error("indexator run failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
