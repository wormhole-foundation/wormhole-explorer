package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// tracer logs queries.
type tracer struct {
	logger *zap.Logger
}

// TraceQueryStart is called when a query is about to start.
func (m *tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	m.logger.Info("query", zap.String("sql", data.SQL), zap.Any("args", data.Args))
	return ctx
}

// TraceQueryEnd is called when a query completes.
func (m *tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

// WithTracer configures a DB to log queries.
func WithTracer(l *zap.Logger) Option {
	return func(d *DB) {
		m := tracer{logger: l}
		d.config.ConnConfig.Tracer = &m
	}
}
