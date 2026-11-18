package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type orgCtxKey struct{}

// OrgCtxKey is the context key under which the caller's organization ID is
// carried. It is exported so the HTTP layer can attach it to the fasthttp
// request context (fasthttp.RequestCtx.Value reads user values).
var OrgCtxKey any = orgCtxKey{}

// WithOrg returns a context scoped to the given organization. Repository
// queries on tenant tables filter by it and stamp it onto inserts.
func WithOrg(ctx context.Context, orgID int64) context.Context {
	return context.WithValue(ctx, OrgCtxKey, orgID)
}

// OrgFromContext returns the organization the context is scoped to. The bool is
// false for system contexts (background jobs, migrations), which are unscoped.
func OrgFromContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	switch v := ctx.Value(OrgCtxKey).(type) {
	case int64:
		return v, true
	case *int64:
		if v != nil {
			return *v, true
		}
	}
	return 0, false
}

// orgArg returns the value to bind for an org-filter parameter: the org ID, or
// nil for an unscoped system context. Pair it with orgFilter in SQL.
func orgArg(ctx context.Context) any {
	if id, ok := OrgFromContext(ctx); ok {
		return id
	}
	return nil
}

// orgInsert returns the organization_id to stamp on an inserted row. System
// contexts fall back to the default organization (1), matching the column default.
func orgInsert(ctx context.Context) int64 {
	if id, ok := OrgFromContext(ctx); ok {
		return id
	}
	return 1
}

// orgFilter renders a SQL predicate restricting col to the org bound at
// placeholder $n; a NULL binding (system context) matches every row.
//
//	"... WHERE " + orgFilter("s.organization_id", 2)
func orgFilter(col string, n int) string {
	return "(" + col + " = $" + itoa(n) + "::bigint OR $" + itoa(n) + "::bigint IS NULL)"
}

// expandOrg rewrites the org tokens in a query and appends the matching
// arguments, so call sites never juggle placeholder numbers:
//
//	@org(col)  -> (col = $N::bigint OR $N::bigint IS NULL), bound to the caller's
//	             org (NULL in a system context, which matches every row)
//	@orgval    -> $N, bound to the org to stamp on an inserted row
func expandOrg(ctx context.Context, query string, args []any) (string, []any) {
	out := make([]any, len(args), len(args)+2)
	copy(out, args)
	for {
		i := strings.Index(query, "@org(")
		if i < 0 {
			break
		}
		j := strings.Index(query[i:], ")")
		if j < 0 {
			break
		}
		col := query[i+len("@org(") : i+j]
		out = append(out, orgArg(ctx))
		query = query[:i] + orgFilter(col, len(out)) + query[i+j+1:]
	}
	for strings.Contains(query, "@orgval") {
		out = append(out, orgInsert(ctx))
		query = strings.Replace(query, "@orgval", "$"+itoa(len(out)), 1)
	}
	return query, out
}

func (r *Repository) sQuery(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	q, a := expandOrg(ctx, query, args)
	return r.db.Query(ctx, q, a...)
}

func (r *Repository) sQueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	q, a := expandOrg(ctx, query, args)
	return r.db.QueryRow(ctx, q, a...)
}

func (r *Repository) sExec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	q, a := expandOrg(ctx, query, args)
	return r.db.Exec(ctx, q, a...)
}

func txExec(ctx context.Context, tx pgx.Tx, query string, args ...any) (pgconn.CommandTag, error) {
	q, a := expandOrg(ctx, query, args)
	return tx.Exec(ctx, q, a...)
}

func txQueryRow(ctx context.Context, tx pgx.Tx, query string, args ...any) pgx.Row {
	q, a := expandOrg(ctx, query, args)
	return tx.QueryRow(ctx, q, a...)
}

func txQuery(ctx context.Context, tx pgx.Tx, query string, args ...any) (pgx.Rows, error) {
	q, a := expandOrg(ctx, query, args)
	return tx.Query(ctx, q, a...)
}

// orgRef names a row that a write points at (a foreign key) so it can be
// checked against the caller's organization before the write goes through.
// shared marks tables whose built-in rows (organization_id NULL) are visible to
// every organization.
type orgRef struct {
	table  string
	id     *int64
	shared bool
}

func ref(table string, id *int64) orgRef { return orgRef{table: table, id: id} }

func refID(table string, id int64) orgRef { return orgRef{table: table, id: &id} }

func sharedRef(table string, id *int64) orgRef { return orgRef{table: table, id: id, shared: true} }

// checkOrgRefs fails with pgx.ErrNoRows-wrapped errors when any referenced row
// is missing or belongs to another organization. Unscoped (system) contexts
// and nil references pass.
func (r *Repository) checkOrgRefs(ctx context.Context, refs ...orgRef) error {
	org, ok := OrgFromContext(ctx)
	if !ok {
		return nil
	}
	for _, rf := range refs {
		if rf.id == nil {
			continue
		}
		q := "SELECT 1 FROM " + rf.table + " WHERE id = $1 AND (organization_id = $2"
		if rf.shared {
			q += " OR organization_id IS NULL"
		}
		q += ")"
		var one int
		if err := r.db.QueryRow(ctx, q, *rf.id, org).Scan(&one); err != nil {
			return fmt.Errorf("referenced %s %d: %w", strings.TrimSuffix(rf.table, "s"), *rf.id, err)
		}
	}
	return nil
}

// firstPtr returns the first non-nil optional ID from a variadic parameter.
func firstPtr(ids []*int64) *int64 {
	if len(ids) > 0 {
		return ids[0]
	}
	return nil
}
