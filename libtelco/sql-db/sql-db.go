// Copyright (C) 2018 Mikhail Masyagin

/*
Package SQLDB содержит "ООП-обертку" (ORM-обертку) над базой данных.
Использование ORM связано с тем, что
*/
package SQLDB

// Виды пользователей: ребенок и родитель.
// Использовать строго enum-константы.
const (
	child  = 0
	parent = 1
)

// User struct содержит основную информацию о пользователе:
// - основной ключ;
// - логин -  у всех записей должен быть различным (!!!);
// - хэш пароля. Вычисляется по правилу:
// md5(salt + md5(password));
// - тип - ребенок или родитель;
// - верифицирован ли аккаунт (email-подтверждение);
// - школа;
// - класс;
// - идентификатор класса;
type User struct {
	ID              int
	Login           string
	HashedPassword  string
	Type            int
	Verified        bool
	School          string
	Class           int8
	ClassIdentifier string
}
