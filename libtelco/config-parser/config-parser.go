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
	ServerAddr     string         `json:"serverAddr"`
	Postgres       string         `json:"postgres"`
	Redis          string         `json:"redis"`
	MaxProcs       int            `json:"maxProcs"`
	PoolSize       int            `json:"poolSize"`
	UpdateInterval int            `json:"updateInterval"`
	LogFile        string         `json:"logFile"`
	SchoolServers  []SchoolServer `json:"schoolServers"`
}

// Типы серверов.
const (
	UnknownType = iota
	FirstType   = iota
	SecondType  = iota
)

// SchoolServer содержит информацию об очередном школьном сервере.
// Так как сервера бывают разные, то по полю Type мы определяем,
// какой именно парсер должен быть применен к серверу.
type SchoolServer struct {
	Type     int    `json:"type"`
	Link     string `json:"link"`
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
