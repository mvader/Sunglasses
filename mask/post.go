package mask

import (
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ObjectType int
type VideoService int

const (
	// Post types
	PostStatus = 1
	PostPhoto  = 2
	PostVideo  = 3
	PostLink   = 4
	Album      = 5
)

// Post model
type Post struct {
	ID       bson.ObjectId   `json:"id" bson:"_id"`
	UserID   bson.ObjectId   `json:"user_id" bson:"user_id"`
	Created  float64         `json:"created" bson:"created"`
	Type     ObjectType      `json:"post_type" bson:"post_type"`
	Likes    float64         `json:"likes" bson:"likes"`
	Comments float64         `json:"comments" bson:"comments"`
	Reported float64         `json:"reported" bson:"reported"`
	Privacy  PrivacySettings `json:"privacy" bson:"privacy"`
	Text     string          `json:"text,omitempty" bson:"text,omitempty"`

	// Video specific fields
	Serivce VideoService `json:"video_service,omitempty" bson:"video_service,omitempty"`
	VideoID string       `json:"video_id,omitempty" bson:"video_id,omitempty"`
	// Also used in link
	Title string `json:"title,omitempty" bson:"title,omitempty"`

	// Photo specific fields
	PhotoURL  string        `json:"photo_url,omitempty" bson:"photo_url,omitempty"`
	Caption   string        `json:"caption,omitempty" bson:"caption,omitempty"`
	AlbumID   bson.ObjectId `json:"album_id,omitempty" bson:"album_id,omitempty"`
	Thumbnail string        `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`

	// Link specific fields
	URL string `json:"link_url,omitempty" bson:"link_url,omitempty"`
}

// NewPost returns a new post instance
func NewPost(t ObjectType, user *User, r *http.Request) *Post {
	p := new(Post)
	p.Type = t
	p.Created = float64(time.Now().Unix())
	p.UserID = user.ID
	p.Privacy = getPostPrivacy(t, r, user)

	return p
}

// CreatePost creates a new post
func CreatePost(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	var (
		postType     = r.PostFormValue("post_type")
		status   int = 200
		response     = make(map[string]interface{})
		user         = GetRequestUser(r, conn, s)
	)

	switch postType {
	/* TODO implement
	case "photo":
		status, response = postPhoto(user, conn, r)
		break
	case "video":
		status, response = postVideo(user, conn, r)
		break
	case "link":
		status, response = postLink(user, conn, r)
		break*/
	default:
		// Default post type is status
		status, response = postStatus(user, conn, r)
	}

	res.JSON(status, response)
}

/* TODO implement
func postPhoto(user *User, conn *Connection, r *http.Request) (int, map[string]interface{}) {

}

func postVideo(user *User, conn *Connection, r *http.Request) (int, map[string]interface{}) {

}

func postLink(user *User, conn *Connection, r *http.Request) (int, map[string]interface{}) {

}
*/

func postStatus(user *User, conn *Connection, r *http.Request) (int, map[string]interface{}) {
	var (
		responseCode int = 200
		statusText       = strings.TrimSpace(r.PostFormValue("post_text"))
		response         = make(map[string]interface{})
	)

	if strlen(statusText) > 0 && strlen(statusText) <= 1500 {
		post := NewPost(PostStatus, user, r)
		post.Text = statusText
		post.Privacy = getPostPrivacy(PostStatus, r, user)
		if err := post.Save(conn); err != nil {

		}
	} else {
		responseCode = 400
		response["error"] = "Invalid status text given"
	}

	return responseCode, response
}

// Save inserts the Post instance if it hasn't been created yet or updates it if it has
func (p *Post) Save(conn *Connection) error {
	if p.ID.Hex() == "" {
		p.ID = bson.NewObjectId()
	}

	if err := conn.Save("posts", p.ID, p); err != nil {
		return err
	}

	return nil
}

func getPostPrivacy(postType ObjectType, r *http.Request, u *User) PrivacySettings {
	p := PrivacySettings{}
	var pType int64
	var err error

	if pType, err = strconv.ParseInt(r.PostFormValue("privacy_type"), 10, 8); err != nil {
		pType = 0
	}

	privacyType := PrivacyType(pType)
	defaultSettings := u.Settings.GetPrivacySettings(postType)
	if privacyType == 0 {
		p.Type = defaultSettings.Type
	} else {
		if isValidPrivacyType(privacyType) {
			p.Type = privacyType
		} else {
			p.Type = defaultSettings.Type
		}
	}

	if p.Type == PrivacyAllBut || p.Type == PrivacyNoneBut {
		if privacyType == 0 {
			p.Users = defaultSettings.Users
		} else {
			us, ok := r.PostForm["privacy_users"]
			if ok && len(us) > 0 {
				p.Users = make([]bson.ObjectId, 0, len(us))
				for _, u := range us {
					if bson.IsObjectIdHex(u) {
						p.Users = append(p.Users, bson.ObjectIdHex(u))
						// TODO check if the users are followed by the user
					}
				}
			}
		}
	}

	return p
}
