// push-delete
package restapi

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/masyagin1998/SchoolServer/libtelco/push"
	db "github.com/masyagin1998/SchoolServer/libtelco/sql-db"
	"github.com/pkg/errors"
)

// pushDelete достает девайсы пользователя и рассылает на них пуш удаления
func (rest *RestAPI) pushDelete(userName string, schoolID int, category string) error {
	var (
		usr     db.User
		devices []db.Device
	)

	// Получаем пользователя по логину и schoolID
	where := db.User{Login: userName, SchoolID: uint(schoolID)}
	err := rest.Db.SchoolServerDB.Where(where).First(&usr).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}

	// Достанем все девайсы пользователя
	err = rest.Db.SchoolServerDB.Model(&usr).Related(&devices).Error
	if err != nil {
		return errors.Wrapf(err, "Error when getting user='%v' devices list", usr)
	}

	// Гоним по девайсам
	for _, dev := range devices {
		err = rest.sendPushDelete(dev.SystemType, dev.Token, category)
		if err != nil {
			return err
		}
	}
	return nil
}

// sendPushDelete посылает push-уведомление по web api gorush с приказом
// удалить уведомления заданной категории
func (rest *RestAPI) sendPushDelete(systemType int, token, category string) error {
	var notifications []push.Notification
	notifications = make([]push.Notification, 1)
	notifications[0] = push.Notification{
		Tokens:           []string{token},
		Platform:         systemType,
		Badge:            1,
		Category:         category,
		ContentAvailable: true,
		Topic:            rest.Push.AppTopic,
		Alert:            push.Alert{Title: "delete"},
	}
	req := push.GorushRequest{Notifications: notifications}
	byt, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "REST: Error marshalling notification")
	}
	resp, err := http.Post(rest.Push.GorushAddress, "application/json", bytes.NewBuffer(byt))
	if err != nil {
		return errors.Wrap(err, "REST: Error sending web api gorush request")
	}
	defer resp.Body.Close()
	return nil
}
