package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"SchoolServer/libtelco/config-parser"
	"strconv"
	"errors"
)

type Database struct  {
	UsersDB *gorm.DB // вас вообще не должно волновать как эти поля называются
	SchoolsDB *gorm.DB
}


type User struct {
	gorm.Model
	Login string
	Password string
	SchoolID int // у гимназии кирилла 1 должно быть
	Permission bool
}

type School struct {
	gorm.Model
	Address string
	Permission bool
}

func NewDatabase() *Database {//создает новую структуру Database и возвращает указатель на неё.
	//предполагается, что базы users и schools уже есть, и в них колонки как поля в User и School
	//вот здесь надо разобраться где мы храним БД и туда всё направить
	udb,err :=gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=users sslmode=disable")
	if err!=nil {
		panic(err)
	}
	//вот здесь надо разобраться где мы храним БД и туда всё направить
	sdb,err :=gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=schools sslmode=disable")
	if err!=nil {
		panic(err)
	}
	return &Database{udb,sdb}
}

func (db *Database) UpdateUser(login string, passkey string, id int/*было стринг но я решил: в  интах проще*/) bool {
	//обновляет данные о пользователе. Если пользователя с таким именем нет в БД, создаёт.
	var user User
	db.UsersDB.Where("login = ?",login).First(&user)
	if user.ID>0 { //существует
		user.Password=passkey
		user.SchoolID=id
		db.UsersDB.Save(&user)
		return true
	}else {
		db.UsersDB.Create(&User{Login:login,Password:passkey,SchoolID:id,Permission:true})
		return false
	}
	//возвращаемое значение:
	//true, если пользователь ранее существовал, false иначе
}

func (db *Database) GetUserAuthData(userName string) (*configParser.School, error){
	//возвращает данные для повторной авторизации пользователя с именем userName
	var school School
	var user User
	db.UsersDB.Where("login = ?",userName).First(&user);
	if user.ID==0 {
		return &configParser.School{},errors.New("User with name "+userName+" doesn't exist")
	}
	db.SchoolsDB.First(&school, user.SchoolID)
	if school.ID==0 {
		return &configParser.School{},errors.New("School with ID "+strconv.Itoa(user.SchoolID)+" doesn't exist")
	}
	return &configParser.School{Link:school.Address,Login:userName,Password:user.Password,},nil
	//*cp.School –  указатель на структуру School из “SchoolServer/libtelco/config-parser“,
	//у которой обязательно заполнены поля Link, Login, Password;
	// где Link – путь к серверу школы пользователя без указания протокола (!) (например, “62.117.74.43”),
	// Login – имя пользователя, Password – md5-хеш пароля пользователя
	//err – ошибка, если она была, nil иначе
}

func (db *Database) GetPermission(userName string, schoolId int/*было стринг но я решил: в  интах проще*/) bool {
	//возвращает permission для пользователя школы с номером schoolId.
	var (
		user User
		school School
	)
	db.UsersDB.Where("login = ?",userName).First(&user)
	db.SchoolsDB.First(&school,schoolId)
	return user.Permission||school.Permission;//ало сань тут точно ИЛИ а не И? 
}
