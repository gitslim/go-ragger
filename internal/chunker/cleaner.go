package chunker

import (
	"context"
	"time"
)

// resetStuckDocuments resets status of documents that are stuck in chunking process for too long.
func (p *Chunker) resetStuckDocuments(ctx context.Context) error {
	err := p.db.ResetStuckChunkingDocuments(ctx)
	return err
}

// stuckDocumentsCleaner periodically run cleaning function
func (p *Chunker) stuckDocumentsCleaner(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.StuckTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.resetStuckDocuments(ctx); err != nil {
				p.logger.Error("failed to reset stuck chunking documents", "error", err)
			}
		}
	}
}
