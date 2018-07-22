// Copyright (C) 2018 Mikhail Masyagin

/*
Package configParser содержит функцию чтения кофигурационного файла сервера.
*/
package configParser

import (
	"encoding/json"
	"os"
)

// Config struct содержит конфигурацию бота.
// Для упрощения чтения конфигов использовано тэгирование.
// - адрес сервера;
// - sql базы данных;
// - inMemory базу данных;
// - максимальное количество потоков;
// - размер пула;
// - время обновления (в секундах);
// - информация о лог-файле;
// - параметры школьных серверов;
type Config struct {
	ServerAddr     string   `json:"serverAddr"`
	Postgres       string   `json:"postgres"`
	MaxProcs       int      `json:"maxProcs"`
	UpdateInterval int      `json:"updateInterval"`
	LogFile        string   `json:"logFile"`
	Schools        []School `json:"schools"`
}

// Типы серверов.
const (
	UnknownType = iota
	FirstType   = iota
	SecondType  = iota
)

// School содержит информацию об очередной школе.
// Так как школьные сервера бывают разные, то по полю Type мы определяем,
// какой именно парсер должен быть применен к серверу.
type School struct {
	Name       string `json:"name"`
	Type       int    `json:"type"`
	Link       string `json:"link"`
	Time       int    `json:"time"`
	Permission bool   `json:"permission"`
	// Этих полей скоро не будет.
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ReadConfig читает конфигурационный файл.
func ReadConfig() (config *Config, err error) {
	configFile, err := os.Open("config.json")
	if err != nil {
		return
	}
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}
	err = configFile.Close()
	return
}
