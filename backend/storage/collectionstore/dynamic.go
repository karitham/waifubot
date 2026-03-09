package collectionstore

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
)

type ListPaginatedParams struct {
	UserID     uint64
	SearchTerm string
	OrderBy    string // "date" or "name" or "id"
	Ascending  bool   // true for ASC, false for DESC
	CursorDate *time.Time
	CursorID   int64
	CursorName string
	Limit      int32
}

// ListPaginatedDynamic builds and executes a dynamic query for paginated character listing
func (q *Queries) ListPaginatedDynamic(ctx context.Context, arg ListPaginatedParams) ([]ListPaginatedRow, error) {
	// Build base query
	query := squirrel.Select(
		"c.id",
		"c.name",
		"c.image",
		"c.media_title",
		"col.source",
		"col.acquired_at AS date",
	).
		From("collection col").
		Join("characters c ON col.character_id = c.id").
		Where(squirrel.Eq{"col.user_id": arg.UserID})

	// Add search filter if provided
	if arg.SearchTerm != "" {
		query = query.Where(
			squirrel.Or{
				squirrel.ILike{"c.name": "%" + arg.SearchTerm + "%"},
				squirrel.Like{"c.id::text": arg.SearchTerm + "%"},
			},
		)
	}

	// Determine sort column and direction
	sortCol := "col.acquired_at"
	tiebreakerCol := "c.id"

	switch arg.OrderBy {
	case "name":
		sortCol = "c.name"
		// For name sorting, we need to handle cursor differently
		if arg.CursorName != "" || arg.CursorID != 0 {
			if arg.Ascending {
				query = query.Where(
					squirrel.Or{
						squirrel.Gt{"c.name": arg.CursorName},
						squirrel.And{
							squirrel.Eq{"c.name": arg.CursorName},
							squirrel.Gt{"c.id": arg.CursorID},
						},
					},
				)
			} else {
				query = query.Where(
					squirrel.Or{
						squirrel.Lt{"c.name": arg.CursorName},
						squirrel.And{
							squirrel.Eq{"c.name": arg.CursorName},
							squirrel.Lt{"c.id": arg.CursorID},
						},
					},
				)
			}
		}
	case "id":
		sortCol = "c.id"
		tiebreakerCol = "" // No tiebreaker needed for ID sorting
		if arg.CursorID != 0 {
			if arg.Ascending {
				query = query.Where(squirrel.Gt{"c.id": arg.CursorID})
			} else {
				query = query.Where(squirrel.Lt{"c.id": arg.CursorID})
			}
		}
	default: // "date"
		if arg.CursorDate != nil && !arg.CursorDate.IsZero() {
			if arg.Ascending {
				query = query.Where(
					squirrel.Or{
						squirrel.Gt{"col.acquired_at": *arg.CursorDate},
						squirrel.And{
							squirrel.Eq{"col.acquired_at": *arg.CursorDate},
							squirrel.Gt{"c.id": arg.CursorID},
						},
					},
				)
			} else {
				query = query.Where(
					squirrel.Or{
						squirrel.Lt{"col.acquired_at": *arg.CursorDate},
						squirrel.And{
							squirrel.Eq{"col.acquired_at": *arg.CursorDate},
							squirrel.Lt{"c.id": arg.CursorID},
						},
					},
				)
			}
		}
	}

	// Add ORDER BY
	if arg.Ascending {
		query = query.OrderBy(sortCol + " ASC")
		if tiebreakerCol != "" {
			query = query.OrderBy(tiebreakerCol + " ASC")
		}
	} else {
		query = query.OrderBy(sortCol + " DESC")
		if tiebreakerCol != "" {
			query = query.OrderBy(tiebreakerCol + " DESC")
		}
	}

	// Add LIMIT
	query = query.Limit(uint64(arg.Limit))

	// Build SQL
	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Execute query
	rows, err := q.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListPaginatedRow
	for rows.Next() {
		var i ListPaginatedRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Image,
			&i.MediaTitle,
			&i.Source,
			&i.Date,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// CountFiltered counts total characters for a user with optional search filter
func (q *Queries) CountFiltered(ctx context.Context, userID uint64, searchTerm string) (int64, error) {
	query := squirrel.Select("COUNT(*)").
		From("collection col").
		Join("characters c ON col.character_id = c.id").
		Where(squirrel.Eq{"col.user_id": userID})

	if searchTerm != "" {
		query = query.Where(
			squirrel.Or{
				squirrel.ILike{"c.name": "%" + searchTerm + "%"},
				squirrel.Like{"c.id::text": searchTerm + "%"},
			},
		)
	}

	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var count int64
	err = q.db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
