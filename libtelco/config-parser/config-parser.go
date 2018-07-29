// Copyright (C) 2018 Mikhail Masyagin

/*
Package configParser содержит функцию чтения кофигурационного файла сервера.
*/
package configParser

import (
	"encoding/json"
	"os"
)

// Config struct содержит конфигурацию сервера.
type Config struct {
	ServerAddr  string    `json:"serverAddr"`
	Postgres    *Postgres `json:"postgres"`
	Redis       *Redis    `json:"redis"`
	CookieStore *Redis    `json:"cookieStore"`
	MaxProcs    int       `json:"maxProcs"`
	LogFile     string    `json:"logFile"`
	Schools     []School  `json:"schools"`
}

// Типы серверов.
const (
	UnknownType = iota
	FirstType   = iota
	SecondType  = iota
)

// Postgres struct содержит конфигурацию PostgreSQL-базы данных.
type Postgres struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBname   string `json:"dbname"`
	SSLmode  string `json:"sslmode"`
}

// Redis struct содержит конфигурацию Redis-базы данных.
type Redis struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DBname   int    `json:"dbname"`
}

// School содержит информацию об очередной школе.
// Так как школьные сервера бывают разные, то по полю Type мы определяем,
// какой именно парсер должен быть применен к серверу.
type School struct {
	Name       string `json:"name"`
	Type       int    `json:"type"`
	Link       string `json:"link"`
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
