package middleware

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/sessions"
	"github.com/martini-contrib/render"
	"github.com/mvader/sunglasses/models"
	"github.com/mvader/sunglasses/modules/auth"
	"github.com/mvader/sunglasses/services"
	. "github.com/mvader/sunglasses/util"
	"labix.org/v2/mgo"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	Config         *services.Config
	Conn           *services.Connection
	Request        *http.Request
	Render         render.Render
	Session        *sessions.Session
	User           *models.User
	Tasks          *services.TaskService
	ResponseWriter http.ResponseWriter
	IsWebToken     bool
}

// CreateContext initializes the context for a request
func CreateContext(ctx martini.Context, config *services.Config, conn *services.Connection, render render.Render, r *http.Request, s *sessions.CookieStore, ts *services.TaskService, rw http.ResponseWriter) {
	c := Context{Config: config, Conn: conn, Request: r, Render: render, ResponseWriter: rw, Tasks: ts}

	if r != nil && s != nil && conn != nil {
		c.Session, _ = s.Get(r, config.SessionName)
		c.User, c.IsWebToken = auth.GetRequestUser(r, conn, c.Session)
	}

	if !c.IsWebToken && r.Header.Get("X-Access-Token") == "" && r.Header.Get("X-User-Token") == "" {
		c.IsWebToken = true
	}

	ctx.Map(c)
}

// Error renders a single error message
func (c Context) Error(status, code int, message string) {
	c.Render.JSON(status, map[string]interface{}{
		"error":   true,
		"single":  true,
		"code":    code,
		"message": message,
	})
}

// Errors renders a json response with an array of errors
func (c Context) Errors(status int, codes []int, messages []string) {
	c.Render.JSON(status, map[string]interface{}{
		"error":    true,
		"single":   false,
		"messages": messages,
		"codes":    codes,
	})
}

// Success renders a successful JSON response
func (c Context) Success(status int, data map[string]interface{}) {
	data["error"] = false
	c.Render.JSON(status, data)
}

// ListCountParams returns the count and offset parameters from the request
func (c Context) ListCountParams() (int, int) {
	var (
		count, offset int64
		err           error
	)

	if count, err = strconv.ParseInt(c.Request.FormValue("count"), 10, 8); err != nil {
		count = 25
	}

	if count > 100 || count < 5 {
		count = 25
	}

	if offset, err = strconv.ParseInt(c.Request.FormValue("offset"), 10, 8); err != nil {
		offset = 0
	}

	return int(count), int(offset)
}

// Form returns the value at the given form key
func (c Context) Form(name string) string {
	return c.Request.FormValue(name)
}

func (c Context) GetBoolean(key string) bool {
	if v := c.Form(key); v != "" {
		if strings.ToLower(v) == "true" || v == "1" {
			return true
		}
	}

	return false
}

// Query returns a pointer to a collection
func (c Context) Query(colName string) *mgo.Collection {
	return c.Conn.Db.C(colName)
}

func (c Context) Find(colName string, where interface{}) *mgo.Query {
	return c.Query(colName).Find(where)
}

func (c Context) FindId(colName string, id interface{}) *mgo.Query {
	return c.Query(colName).FindId(id)
}

func (c Context) Remove(colName string, selector interface{}) error {
	return c.Query(colName).Remove(selector)
}

func (c Context) RemoveAll(colName string, selector interface{}) (*mgo.ChangeInfo, error) {
	return c.Query(colName).RemoveAll(selector)
}

func (c Context) Count(colName string, query interface{}) (int, error) {
	return c.Find(colName, query).Count()
}

func (c Context) AsyncQuery(fn func(*services.Connection)) {
	conn := new(services.Connection)
	conn.Session = c.Conn.Session.Copy()
	conn.Db = conn.Session.DB(c.Config.DatabaseName)

	fn(conn)
}

// RequestIsValid returns if the current request signature is valid and thus is a valid request
func (c Context) RequestIsValid(isAccessKey bool) bool {
	signature := c.Form("signature")
	URL := c.Request.URL

	if signature != "" {
		timestamp, err := strconv.ParseInt(c.Form("timestamp"), 10, 64)
		if err != nil || time.Now().Unix()-timestamp > 300 {
			return false
		}

		isAPIRequest := c.Request.Header.Get("X-User-Token") != "" || isAccessKey
		key := c.Form("api_key")

		if isAPIRequest {
			return validateAPISignature(c.Conn, signature, timestamp, key, URL)
		} else {
			var csrfKey string
			if c.Session.Values["csrf_token"] == nil {
				csrfKey = ""
			} else {
				csrfKey = c.Session.Values["csrf_token"].(string)
			}

			if csrfKey == "" {
				return false
			}

			return validateWebSignature(signature, timestamp, csrfKey, URL)
		}
	}

	return false
}

func validateAPISignature(conn *services.Connection, signature string, timestamp int64, key string, URL *url.URL) bool {
	/* TODO Application not implemented yet
	var app Application
	if err := conn.Db.C("applications").Find(bson.M{"public_key":Hash(key), "active":true}).One(&app); err != nil {
		return false
	}
	privateKey := app.PrivateKey*/
	privateKey := ""
	return signature == HashMD5(URL.Path+privateKey+fmt.Sprint(timestamp))
}

func validateWebSignature(signature string, timestamp int64, csrfKey string, URL *url.URL) bool {
	return signature == HashMD5(URL.Path+csrfKey+fmt.Sprint(timestamp))
}
