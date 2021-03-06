package handlers

import (
	. "github.com/mvader/sunglasses/error"
	"github.com/mvader/sunglasses/middleware"
	"github.com/mvader/sunglasses/models"
	"labix.org/v2/mgo/bson"
)

// MarkNotificationRead marks a notification as read
func MarkNotificationRead(c middleware.Context) {
	var n models.Notification

	nid := c.Form("notification_id")

	if nid == "" || !bson.IsObjectIdHex(nid) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	notificationID := bson.ObjectIdHex(nid)

	if err := c.FindId("notifications", notificationID).One(&n); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	if n.User != c.User.ID {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	if !n.Read {
		n.Read = true

		if _, err := c.Query("notifications").UpsertId(n.ID, n); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}
	}

	c.Success(200, map[string]interface{}{
		"message": "Notification marked successfully as read",
	})
}

// ListNotifications list all the user's notifications
func ListNotifications(c middleware.Context) {
	count, offset := c.ListCountParams()
	var result models.Notification
	notifications := make([]models.Notification, 0, count)

	cursor := c.Find("notifications", bson.M{"user_id": c.User.ID}).Sort("-time").Limit(count).Skip(offset).Iter()
	for cursor.Next(&result) {
		notifications = append(notifications, result)
	}

	if err := cursor.Close(); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	users := make([]bson.ObjectId, 0, len(notifications))
	for _, n := range notifications {
		if n.UserActionID.Hex() != "" {
			users = append(users, n.UserActionID)
		}
	}

	usersData := models.GetUsersData(users, c.User, c.Conn)
	if usersData == nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	for i, n := range notifications {
		if u, ok := usersData[n.UserActionID]; ok {
			notifications[i].UserAction = u
		}
	}

	c.Success(200, map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
	return
}
