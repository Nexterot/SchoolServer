package db

import  (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"SchoolServer/libtelco/config-parser"
	"strconv"
	"errors"
	"fmt"
	"SchoolServer/libtelco/log"
)

type Database struct  {
	SchoolServerDB *gorm.DB //указатель на гормовскую дб
	Logger *log.Logger
}


type User struct {
	gorm.Model
	Login string //логин пользователя
	Password string //зашифрованный пароль пользователя
	SchoolID int //ид школы как в дб для школ
	Permission bool //разрешение
}

type School struct {
	gorm.Model
	Name		string //название школы
	Address 	string //интернет адрес сервера
	Permission 	bool //разрешение
}

type school struct {
	Name 	string `json:"name"`
	Id 		string `json:"id"`
	Website string `json:"website"`
}

type SchoolList struct {
	Schools []school `json:"schools"`
}


func NewDatabase(log* log.Logger) (*Database,error) {//создает новую структуру Database и возвращает указатель на неё.
	//открываем дб
	sdb,err :=gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=schoolserverdb sslmode=disable")
	log.Info("Connecting to DB")
	if err!=nil {
		log.Error(err.Error())
		return &Database{},errors.New("Can't connect to database")
	}
	if !sdb.HasTable(&User{}){ //добавляем таблицу с пользователями, если её нет
		err:=sdb.CreateTable(&User{}).Error
		log.Info("Creating 'users' table")
		if err!=nil {
			log.Error(err.Error())
			return &Database{},err
		}
	}

	if !sdb.HasTable(&School{}){ //добавляем таблицу со школами, если её нет
		err:=sdb.CreateTable(&School{}).Error
		log.Info("Creating 'schools' table")
		if err!=nil {
			log.Error(err.Error())
			return &Database{},err
		}
	}
	return &Database{SchoolServerDB:sdb,Logger:log},nil
}

func (db *Database) UpdateUser(login string, passkey string, id int) (bool,error) {
	//обновляет данные о пользователе. Если пользователя с таким именем нет в БД, создаёт.
	var user User
	err:=db.SchoolServerDB.Find(&User{}).Where("login = ?",login).First(&user).Error //ищем пользователя в дб по логину
	if err!=nil {
		db.Logger.Info(err.Error())
		return false,err
	}
	db.Logger.Info("Finding user ",login," in DB\n")
	if user.ID>0 { //существует
		user.Password=passkey
		user.SchoolID=id
		err=db.SchoolServerDB.Save(&user).Error //обновляем данные
		db.Logger.Info("Updating user ",login," in DB\n")
		if err!=nil {
			db.Logger.Info(err.Error())
			return false,err
		}
		return true,nil
	}else { //не существует
		err=db.SchoolServerDB.Create(&User{Login:login,Password:passkey,SchoolID:id,Permission:true}).Error
		db.Logger.Info("Creating user ",login," in DB\n")
		if err!=nil {
			db.Logger.Error(err.Error())
			return false,err
		}
		return false,nil
	}
	//возвращаемое значение:
	//true, если пользователь ранее существовал, false иначе
}

func (db *Database) GetUserAuthData(userName string) (*configParser.School, error){
	//возвращает данные для повторной авторизации пользователя с именем userName
	var school School
	var user User
	err:=db.SchoolServerDB.Find(&User{}).Where("login = ?",userName).First(&user).Error //ищем пользователя по логину
	db.Logger.Info("Finding user ",userName,"in DB\n")
	if err!=nil {
		db.Logger.Error(err.Error())
		return &configParser.School{},err
	}
	if user.ID==0 { //нет такого пользователя
		db.Logger.Error("User with name "+userName+" doesn't exist");
		return &configParser.School{},fmt.Errorf("User with name %s doesn't exist",userName)
	}
	err=db.SchoolServerDB.Find(&School{}).First(&school, user.SchoolID).Error //ищем школу пользователя по ид
	db.Logger.Info("Finding school ",user.SchoolID," in DB");
	if err!=nil {
		db.Logger.Error(err.Error())
		return &configParser.School{},err
	}
	if school.ID==0 {
		db.Logger.Error("School with ID ",user.SchoolID," doesn't exist")
		return &configParser.School{},errors.New("School with ID "+strconv.Itoa(user.SchoolID)+" doesn't exist")
	}
	return &configParser.School{Link:school.Address,Login:userName,Password:user.Password,},nil
	//*cp.School –  указатель на структуру School из “SchoolServer/libtelco/config-parser“,
	//у которой обязательно заполнены поля Link, Login, Password;
	// где Link – путь к серверу школы пользователя без указания протокола (!) (например, “62.117.74.43”),
	// Login – имя пользователя, Password – md5-хеш пароля пользователя
	//err – ошибка, если она была, nil иначе
}

func (db *Database) GetPermission(userName string, schoolId int,log* log.Logger) (bool,error) {
	//возвращает permission для пользователя школы с номером schoolId.
	var (
		user User
		school School
	)
	err:=db.SchoolServerDB.Find(&User{}).Where("login = ?",userName).First(&user).Error //ищем пользователя по логину
	db.Logger.Info("Finding user %s in DB",userName)
	if err!=nil {
		db.Logger.Error(err.Error())
		return false,err
	}
	err=db.SchoolServerDB.Find(&School{}).First(&school,schoolId).Error //ищем школу по ид
	db.Logger.Info("Finding school %d in DB",schoolId)
	if err!=nil {
		db.Logger.Error(err.Error())
		return false,err
	}
	return user.Permission||school.Permission,nil;
}

//возвращает информацию о всех поддерживаемых школах.
//возвращаемые значения:
//*SchoolList представляет собой указатель на заполненную json-сериализуемую
//структуру, соответствующую структуре ответа на 1.1 Запрос списка школ из ТЗ.
//error – объект ошибки, если она была, nil иначе
func (db *Database) GetSchools() (*SchoolList, error) {
	schDB:=db.SchoolServerDB.Find(&School{})
	db.Logger.Info("Getting schools table")
	err:=schDB.Error
	if err!=nil {
		db.Logger.Error(err.Error())
		return &SchoolList{},err
	}
	sz:=0
	db.Logger.Info("Counting table size")
	for i:=1;;i++{
		err=db.SchoolServerDB.First(&School{},i).Error
		if err!=nil {
			if err.Error()=="record not found"{
				sz=i-1
				break
			} else {
				db.Logger.Error(err.Error())
				return &SchoolList{},err
			}
		}
	}
	db.Logger.Info("Table size is equal to",sz)
	var SList SchoolList
	var s School
	SList.Schools=make([]school,sz)
	schTable:=db.SchoolServerDB.Find(&School{})
	err=schTable.Error
	if err!=nil {
		db.Logger.Error(err.Error())
		return &SchoolList{},err
	}
	for i:=1;i<=sz;i++{
		s.ID=uint(i)
		err=schTable.Where("id = ?", i).First(&s).Error
		db.Logger.Info("Getting school with id ",i)
		if err!=nil {
			db.Logger.Error(err.Error())
			return &SchoolList{},err
		}
		SList.Schools[i-1]=school{Name:s.Name,Id:strconv.Itoa(int(s.ID)),Website:s.Address}
	}
	return &SList,nil
	}
