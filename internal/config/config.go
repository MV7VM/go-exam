// Package config - пакет для работы с конфигурацией.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
)

const (
	defaultCountWorkers   = 10
	defaultSizeChannel    = 100
	defaultPeriodPolling  = 2
	defaultPollingBot     = 10
	defaultContextTimeout = 1
)

// Config Конфигурация приложения.
type Config struct {
	TelegramToken      string   `json:"telegramToken"`
	CountWorkers       int      `json:"workersCount"`
	SizeChannel        int      `json:"optimalChanSize"`
	PeriodPolling      int      `json:"periodPolling"`
	SalutSpeech        *Setting `json:"salutSpeech"`
	GigaChat           *Setting `json:"gigachat"`
	DSN                string   `json:"dsn"`
	PollingBot         int      `json:"pollingBot"`
	ContextTimeout     int      `json:"contextTimeout"`
	DirectoryLoadAudio string   `json:"dirLoadAudio"`
}

type Setting struct {
	Token string `json:"token"`
	Auth  string `json:"authHost"`
	Main  string `json:"requestHost"`
}

func NewConfig() (*Config, error) {

	// Поддерживаем и латинскую "-c", и кириллическую "-с" (частая опечатка на RU раскладке).
	// Важно: обязательно парсим флаги, иначе значения никогда не установятся.
	fileConfigFlagLatin := flag.String("c", "config.json", "Файл конфигурации")
	fileConfigFlagCyrillic := flag.String("с", "", "Файл конфигурации (кириллическая 'с')")
	flag.Parse()

	configPath := *fileConfigFlagLatin
	if *fileConfigFlagCyrillic != "" {
		configPath = *fileConfigFlagCyrillic
	}

	configText, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("невозможно прочитать файл %s:%s", configPath, err.Error())
	}

	cfg := &Config{}

	err = json.Unmarshal(configText, &cfg)
	if err != nil {
		return nil, fmt.Errorf("некорректный формат файла:%s", err.Error())
	}
	if cfg.CountWorkers == 0 {
		cfg.CountWorkers = defaultCountWorkers
	}
	if cfg.SizeChannel == 0 {
		cfg.SizeChannel = defaultSizeChannel
	}

	if cfg.PeriodPolling == 0 {
		cfg.PeriodPolling = defaultPeriodPolling
	}

	if cfg.PollingBot == 0 {
		cfg.PollingBot = defaultPollingBot
	}

	if cfg.ContextTimeout == 0 {
		cfg.ContextTimeout = defaultContextTimeout
	}

	if cfg.TelegramToken == "" {
		return nil, errors.New("не задан токен для telegram-bot")
	}

	if cfg.SalutSpeech == nil {
		return nil, errors.New("не задан конфигурация для SalutSpeech")
	}

	if cfg.GigaChat == nil {
		return nil, errors.New("не задан конфигурация для GigaChat")
	}

	if cfg.DSN == "" {
		return nil, errors.New("не задан конфигурация для подключения к БД")
	}

	if cfg.DirectoryLoadAudio == "" {
		return nil, errors.New("не задан директория для загрузки аудио")
	}

	return cfg, nil
}
