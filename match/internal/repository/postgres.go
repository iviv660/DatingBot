package repository

import (
	"context"
	"database/sql"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(db *sql.DB) *PostgresDB {
	return &PostgresDB{db: db}
}

//Like(ctx context.Context, fromUser, toUser int64, isLike bool) error
//CheckMatch(ctx context.Context, user1, user2 int64) (bool, error)
//TodayLikedIDs(ctx context.Context, fromUser int64) ([]int64, error)

func (p *PostgresDB) Like(ctx context.Context, fromUser, toUser int64, isLike bool) error {
	query := `
		INSERT INTO matches (from_user, to_user, is_like)
		VALUES ($1, $2, $3)
		ON CONFLICT (from_user, to_user)
		DO UPDATE SET is_like = EXCLUDED.is_like, created_at = now()
	`
	_, err := p.db.ExecContext(ctx, query, fromUser, toUser, isLike)
	return err
}

// Проверяем взаимный лайк
func (p *PostgresDB) CheckMatch(ctx context.Context, user1, user2 int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM matches m1
			JOIN matches m2 
			  ON m1.from_user = m2.to_user 
			 AND m1.to_user   = m2.from_user
			WHERE m1.from_user = $1 
			  AND m1.to_user   = $2
			  AND m1.is_like   = TRUE
			  AND m2.is_like   = TRUE
		)
	`
	var exists bool
	err := p.db.QueryRowContext(ctx, query, user1, user2).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *PostgresDB) TodayLikedIDs(ctx context.Context, fromUser int64) ([]int64, error) {
	query := `
		SELECT to_user
		FROM matches
		WHERE from_user = $1
		  AND is_like = TRUE
		  AND (created_at AT TIME ZONE 'UTC')::date = (now() AT TIME ZONE 'UTC')::date
	`
	rows, err := p.db.QueryContext(ctx, query, fromUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
