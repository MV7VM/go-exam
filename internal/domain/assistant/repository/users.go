package repository

import (
	"assistant/internal/domain/assistant/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var insertUser string = `
		INSERT INTO recognition.users (id, chat_id, username)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING;
	`
var selectAudioForUser string = `SELECT id, path FROM recognition.users_audio WHERE user_id = $1 and status='DONE' and result_short is not null`

// CreateUser - добавление в users.
func (r *Repository) CreateUser(ctx context.Context, id, chatid int64, username string) error {

	_, err := r.db.Exec(ctx, insertUser, id, chatid, username)
	if err != nil {
		return fmt.Errorf("ошибка вставки в таблицу users: %w", err)
	}

	return nil
}

// GetAudioListForUser -
func (r *Repository) GetAudioListForUser(ctx context.Context, userID int64) ([]*model.AudioShort, error) {
	//var audioList []*model.AudioShort

	rows, err := r.db.Query(ctx, selectAudioForUser, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	audioList, err := pgx.CollectRows(rows, pgx.RowToStructByName[*model.AudioShort])
	if err != nil {
		return nil, err
	}

	//for rows.Next() {
	//	var audio model.AudioShort
	//	err := rows.Scan(
	//		&audio.ID,
	//		&audio.Path,
	//	)
	//	if err != nil {
	//		return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
	//	}
	//	audioList = append(audioList, &audio)
	//}
	//
	//if err = rows.Err(); err != nil {
	//	return nil, fmt.Errorf("ошибка при чтении строк: %w", err)
	//}

	return audioList, nil
}
