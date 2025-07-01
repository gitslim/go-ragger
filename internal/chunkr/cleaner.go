package chunkr

import (
	"context"
	"time"
)

func (p *DocumentProcessor) resetStuckDocuments(ctx context.Context) error {
	err := p.db.ResetStuckDocuments(ctx)
	return err
}

func (p *DocumentProcessor) stuckDocumentsCleaner(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.StuckTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.resetStuckDocuments(ctx); err != nil {
				p.logger.Error("failed to reset stuck documents", "error", err)
			}
		}
	}
}
