package salutspeech

import (
	"assistant/internal/config"
	"assistant/internal/domain/assistant/auth"
	"assistant/internal/domain/assistant/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	endpointsUpload    = "/data:upload"
	endpointsRecognize = "/speech:async_recognize"
	endpointsGetStatus = "/task:get"
	endpointsGetData   = "/data:download"
)

type Client struct {
	client   *resty.Client
	authHost string
	mainHost string
	authKey  string

	tokenManager *auth.TokenManager

	requests chan *model.Task

	lgr           *zap.Logger
	periodPolling time.Duration
}

type ResultID struct {
	FileID string `json:"request_file_id"`
}

type ResponseUploadFile struct {
	Status int       `json:"status"`
	Result *ResultID `json:"result"`
}

type RequestOptions struct {
	Model         string `json:"model"`
	Encoding      string `json:"audio_encoding"`
	SampleRate    int    `json:"sample_rate,omitempty"`
	ChannelsCount int    `json:"channels_count,omitempty"`
	Language      string `json:"language"`
}

type RequestRecognize struct {
	Options *RequestOptions `json:"options"`
	FileID  string          `json:"request_file_id"`
}

type StatusTask struct {
	ID             string `json:"id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	Status         string `json:"status"`
	ResponseFileID string `json:"response_file_id,omitempty"`
	MsgError       string `json:"error,omitempty"`
}

type ResponseRecognize struct {
	Status int         `json:"status"`
	Result *StatusTask `json:"result"`
}

func NewSalutSpeechClient(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Client, error) {
	//cfg.SalutSpeech, cfg.CountWorkers, cfg.SizeChannel, cfg.PeriodPolling
	//setting *config.Setting, countWorkers, sizeChanel, periodPolling int
	salutSpeech := &Client{
		client:        resty.New(),
		authHost:      cfg.SalutSpeech.Auth,
		authKey:       cfg.SalutSpeech.Token,
		mainHost:      cfg.SalutSpeech.Main,
		requests:      make(chan *model.Task, cfg.SizeChannel),
		lgr:           logger,
		periodPolling: time.Duration(cfg.PeriodPolling) * time.Second,
	}

	salutSpeech.tokenManager = auth.NewTokenManager(salutSpeech.fetchToken)

	salutSpeech.StartWorkers(ctx, cfg.CountWorkers)

	return salutSpeech, nil
}

func (ss *Client) fetchToken() (*model.Token, error) {
	rqUID := model.NewUUID()

	response, err := ss.client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("RqUID", rqUID).
		SetHeader("Authorization", "Basic "+ss.authKey).
		SetFormData(map[string]string{
			"scope": "SALUTE_SPEECH_PERS",
		}).
		Post(ss.authHost)
	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s", response.Body())
	}

	token := &model.Token{}
	if err := json.Unmarshal(response.Body(), token); err != nil {
		return nil, fmt.Errorf("некорректный формат ответа: %w, body: %s", err, response.String())
	}

	return token, nil
}

func (ss *Client) DoWithRetry(req *resty.Request, method, url string) (*resty.Response, error) {
	return ss.tokenManager.DoWithRetry(req, method, url)
}

func (ss *Client) GetCurrentToken() string {
	return ss.tokenManager.CurrentToken()
}

func (ss *Client) UploadFile(audioFilePath string, contentType string) (string, error) {

	response, err := ss.DoWithRetry(ss.client.R().
		SetHeader("Content-Type", contentType).
		SetFile("audio", audioFilePath).
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+ss.GetCurrentToken()),
		"POST",
		ss.mainHost+endpointsUpload)

	if err != nil {
		return "", err
	}

	if response.StatusCode() != http.StatusOK {
		ss.lgr.Error("ошибка запроса UploadFile", zap.Int("StatusCode", response.StatusCode()), zap.String("Bode", response.String()))
		return "", getErr(response.StatusCode())
	}

	responseUpload := &ResponseUploadFile{}

	err = json.Unmarshal(response.Body(), &responseUpload)
	if err != nil {
		return "", fmt.Errorf("Некорректный формат ответа: %s, body: %s", err.Error(), response.String())
	}

	if responseUpload.Status == 200 {
		return responseUpload.Result.FileID, nil
	} else {
		return "", fmt.Errorf("некорректный ответ body: %s", response.String())
	}

}

func (ss *Client) CreateTaskRecognize(fileID string, encoding string, channels int, sampleRate int) (*StatusTask, error) {

	options := &RequestOptions{Model: "general",
		Language: "ru-RU",
		Encoding: encoding,
	}

	if channels > 0 {
		options.ChannelsCount = channels
	}

	if sampleRate > 0 {
		options.SampleRate = sampleRate
	}

	request := &RequestRecognize{
		Options: options,
		FileID:  fileID,
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("некорректный формат запроса %s", err.Error())
	}

	response, err := ss.DoWithRetry(ss.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+ss.GetCurrentToken()).
		SetBody(requestJson),
		"POST",
		ss.mainHost+endpointsRecognize)

	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		ss.lgr.Error("ошибка запроса CreateTaskRecognize", zap.Int("StatusCode", response.StatusCode()), zap.String("Bode", response.String()))
		return nil, getErr(response.StatusCode())
	}

	responseRecognize := &ResponseRecognize{}

	err = json.Unmarshal(response.Body(), &responseRecognize)
	if err != nil {
		return nil, fmt.Errorf("Некорректный формат ответа: %s, body: %s", err, response.String())
	}

	return responseRecognize.Result, nil

}

func (ss *Client) GetStatusTask(taskID string) (string, string, string, error) {

	response, err := ss.DoWithRetry(ss.client.R().
		SetHeader("Accept", "application/octet-stream").
		SetHeader("Authorization", "Bearer "+ss.GetCurrentToken()).
		SetQueryParams(map[string]string{
			"id": taskID,
		}),
		"GET",
		ss.mainHost+endpointsGetStatus)

	if err != nil {
		return "", "", "", err
	}

	if response.StatusCode() != http.StatusOK {
		ss.lgr.Error("ошибка запроса GetStatusTask", zap.Int("StatusCode", response.StatusCode()), zap.String("Bode", response.String()))
		return "", "", "", getErr(response.StatusCode())
	}

	responseStatus := &ResponseRecognize{}

	err = json.Unmarshal(response.Body(), &responseStatus)
	if err != nil {
		return "", "", "", fmt.Errorf("Некорректный формат ответа: %s, body: %s", err, response.String())
	}

	return responseStatus.Result.Status, responseStatus.Result.ResponseFileID, responseStatus.Result.MsgError, nil

}

func (ss *Client) GetData(fileID string) (string, error) {

	response, err := ss.DoWithRetry(ss.client.R().
		SetHeader("Accept", "application/octet-stream").
		SetHeader("Authorization", "Bearer "+ss.GetCurrentToken()).
		SetQueryParams(map[string]string{
			"response_file_id": fileID,
		}),
		"GET",
		ss.mainHost+endpointsGetData)

	if err != nil {
		return "", err
	}

	if response.StatusCode() != http.StatusOK {
		ss.lgr.Error("ошибка запроса GetData", zap.Int("StatusCode", response.StatusCode()), zap.String("Bode", response.String()))
		return "", getErr(response.StatusCode())
	}

	err = saveBinaryResult(response.Body(), "./result.txt")
	if err != nil {
		ss.lgr.Error(err.Error())
	}

	return response.String(), nil

}

func saveBinaryResult(data []byte, path string) error {
	return os.WriteFile(path, data, 0644)
}

func getErr(statusCode int) error {

	switch statusCode {
	case 400:
		return errors.New("некорректный формат запроса")
	case 401:
		return errors.New("ошибка аторизации")
	case 413:
		return errors.New("превышен максимальный размер входных данных.")
	default:
		return errors.New("Internal Server Error")
	}
}
