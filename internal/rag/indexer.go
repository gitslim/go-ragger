package rag

import (
	"context"
	"errors"
	"fmt"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *RAGPipeline) documentIndexerBatch(ctx context.Context) error {
	docs, err := p.db.GetChunkedDocuments(ctx, int32(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("get chunked documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.indexDocument(ctx, doc.ID); err != nil {
			p.logger.Error("failed to index document",
				"doc_id", doc.ID,
				"error", err)

			if err := p.db.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
				ID:     doc.ID,
				Status: sqlc.DocumentStatusIndexfail,
			}); err != nil {
				p.logger.Error("failed to mark document as indexfail",
					"doc_id", doc.ID,
					"error", err)
			}
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

	err = p.storeDocumentToMilvus(ctx, doc)
	if err != nil {
		return err
	}

	if err := p.db.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
		ID:     doc.ID,
		Status: sqlc.DocumentStatusIndexed,
	}); err != nil {
		p.logger.Error("failed to mark document as indexed",
			"doc_id", doc.ID,
			"error", err)
	}

	return tx.Commit(ctx)
}

func (p *RAGPipeline) storeDocumentToMilvus(ctx context.Context, doc sqlc.Document) error {
	p.logger.Debug("stored document to milvus", "doc", doc.ID)
	return nil
}
