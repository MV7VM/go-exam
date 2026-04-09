package repository

import (
	"assistant/internal/domain/assistant/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

const qGetAudioByID = `SELECT result_short, result_text, is_voice FROM assistant.users_audio WHERE user_id = $1 and id=$2`

func (r *Repository) GetAudioByID(ctx context.Context, userID int64, audioID string) (string, string, bool, error) {
	var text, shortText string
	var isVoice bool

	err := r.db.QueryRow(ctx, qGetAudioByID, userID, audioID).Scan(&shortText, &text, &isVoice)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", false, errors.New("data not found")
		}
		return "", "", false, fmt.Errorf("ошибка получения текст аудио из БД: %w", err)
	}

	return shortText, text, isVoice, nil
}

const qGetAudioByWord = `SELECT a.id, a.path FROM assistant.users_audio a inner join assistant.users_audio_worr w on w.id=a.id WHERE a.user_id = $1 and word=$2 `

func (r *Repository) GetAudioByWord(ctx context.Context, userID int64, word string) ([]*model.AudioShort, error) {
	//var audioList []*model.AudioShort

	rows, err := r.db.Query(ctx, qGetAudioByWord, userID, word)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	audioList, err := pgx.CollectRows(rows, pgx.RowToStructByName[*model.AudioShort])
	if err != nil {
		return nil, err
	}

	return audioList, nil
}

const qCreateAudio = `INSERT INTO assistant.users_audio(id, user_id, path, is_voice) values($1,$2,$3,$4)`

func (r *Repository) CreateAudio(ctx context.Context, userID int64, file string, isVoice bool) (string, error) {
	id := model.NewUUID()

	_, err := r.db.Exec(ctx, qCreateAudio, id, userID, file, isVoice)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения аудио: %w", err)
	}

	return id, nil
}

const qUpdateStatusTask = `update assistant.users_audio 
								set  file_id=$2,task_id=$3,status=$4,updated_at=NOW()
								where id=$1`

func (r *Repository) UpdateStatusTask(ctx context.Context, id, taskID, fileID, status string) error {

	_, err := r.db.Exec(ctx, qUpdateStatusTask, id, toNullString(fileID), toNullString(taskID), status)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса аудио: %w", err)
	}

	return nil
}

const qUpdateShortText = `update assistant.users_audio 
								set  result_short=$2,updated_at=NOW()
								where id=$1`

func (r *Repository) UpdateShortText(ctx context.Context, id string, text string) error {

	_, err := r.db.Exec(ctx, qUpdateShortText, id, text)
	if err != nil {
		return fmt.Errorf("ошибка обновления краткой выжимки аудио: %w", err)
	}

	return nil
}

const (
	qUpdateResult = `update assistant.users_audio 
								set  result=$2, result_text=$3, updated_at=NOW()
								where id=$1`
	qInsertWord string = `
				INSERT INTO assistant.users_audio_worr (id, word)
				VALUES %s
			`
)

func (r *Repository) SaveResult(ctx context.Context, id string, resultJson string, result *model.CombinedResult) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	res, err := tx.Exec(ctx, qUpdateResult, id, resultJson, result.NormalizedText)
	if err != nil {
		return fmt.Errorf("обновление результата выполнено с ошибкой: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("row with id=%s not found", id)
	}

	if len(result.WordAlignments) > 0 {
		valueParts := make([]string, 0, len(result.WordAlignments))
		args := make([]any, 0, len(result.WordAlignments)*2)

		for _, wa := range result.WordAlignments {
			word := strings.TrimSpace(wa.Word)
			if word == "" {
				continue
			}

			valueParts = append(valueParts, "(?, ?)")
			args = append(args, id, word)
		}

		if len(valueParts) > 0 {
			insertQuery := fmt.Sprintf(qInsertWord, strings.Join(valueParts, ","))
			//insertQuery = tx.Rebind(insertQuery)

			if _, err = tx.Exec(ctx, insertQuery, args...); err != nil {
				return fmt.Errorf("insert word: %w", err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
