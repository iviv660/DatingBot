package repository

import (
	"app/user/internal/dto"
	"app/user/internal/entity"
	"context"
	"database/sql"
	"errors"
)

// UserRepo — интерфейс теперь с контекстами, можешь вынести его в doma

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(db *sql.DB) *PostgresDB {
	return &PostgresDB{DB: db}
}

func (db *PostgresDB) RegisterUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (
			telegram_id, username, age, 
			gender, location, description, 
		    photo_url, is_visible, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id
		  `
	err := db.DB.QueryRowContext(
		ctx,
		query,
		user.TelegramID,
		user.Username,
		user.Age,
		user.Gender,
		user.Location,
		user.Description,
		user.PhotoURL,
		user.IsVisible,
		user.CreatedAt,
	).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *PostgresDB) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error) {
	query := `
		SELECT 
			id, telegram_id, username, age,
			gender, location, description,
			photo_url, is_visible, created_at
		FROM users
		WHERE telegram_id = $1
	`
	var description sql.NullString
	var photoURL sql.NullString

	user := &entity.User{}

	err := db.DB.QueryRowContext(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Age,
		&user.Gender,
		&user.Location,
		&description,
		&photoURL,
		&user.IsVisible,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (db *PostgresDB) GetProfile(ctx context.Context, userID int64) (*entity.User, error) {
	query := `
		SELECT 
			id, telegram_id, username, age,
			gender, location, description,
			photo_url, is_visible, created_at
		FROM users
		WHERE id = $1
	`

	var description sql.NullString
	var photoURL sql.NullString

	user := &entity.User{}
	err := db.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Age,
		&user.Gender,
		&user.Location,
		&description,
		&photoURL,
		&user.IsVisible,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (db *PostgresDB) UpdateProfile(ctx context.Context, userID int64, input dto.UpdateProfileInput) (*entity.User, error) {
	query := `
		UPDATE users
		SET username = $1,
			age = $2,
			gender = $3,
			location = $4,
			description = $5
		WHERE id = $6
		RETURNING id, telegram_id, username, age, gender, location, description, photo_url, is_visible, created_at
	`

	var description sql.NullString
	var photoURL sql.NullString

	user := &entity.User{}

	err := db.DB.QueryRowContext(ctx, query,
		input.Username,
		input.Age,
		input.Gender,
		input.Location,
		input.Description,
		userID,
	).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Age,
		&user.Gender,
		&user.Location,
		&description,
		&photoURL,
		&user.IsVisible,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if description.Valid {
		user.Description = description.String
	}
	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}

	return user, nil
}

func (db *PostgresDB) GetCandidates(ctx context.Context, filter dto.CandidateFilter) ([]int64, error) {
	query := "SELECT id FROM users WHERE gender=$1 AND age >= $2 AND age <= $3 AND location=$4 AND is_visible=TRUE LIMIT $5"
	rows, err := db.DB.QueryContext(ctx, query, filter.TargetGender, filter.MinAge, filter.MaxAge, filter.Location, filter.Limit)
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
	return ids, nil
}

func (db *PostgresDB) ToggleVisibility(ctx context.Context, userID int64, isVisible bool) error {
	query := `
		UPDATE users
		SET is_visible = $1
		WHERE id = $2`
	_, err := db.DB.ExecContext(ctx, query, isVisible, userID)
	if err != nil {
		return err
	}
	return nil
}
