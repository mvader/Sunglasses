package mask

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPostStatus(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
	}()

	Convey("Posting a status", t, func() {
		Convey("When no user is passed", func() {
			testPostHandler(CreatePost, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When the status text is invalid", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", randomString(3000))
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidStatusText)
				So(errResp.Message, ShouldEqual, MsgInvalidStatusText)
			})
		})

		Convey("When everything is OK", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", "A test status")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "Status posted successfully")
			})
		})
	})
}

func TestPostVideo(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
	}()

	Convey("Posting a video", t, func() {
		Convey("When no user is passed", func() {
			testPostHandler(CreatePost, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When the status text is invalid", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", randomString(3000))
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidStatusText)
				So(errResp.Message, ShouldEqual, MsgInvalidStatusText)
			})
		})

		Convey("When the link is not valid (youtube)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://youtube.com/watch?v=notfoundvideo")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidVideoURL)
				So(errResp.Message, ShouldEqual, MsgInvalidVideoURL)
			})
		})

		Convey("When the link is not valid (vimeo)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://vimeo.com/00000")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidVideoURL)
				So(errResp.Message, ShouldEqual, MsgInvalidVideoURL)
			})
		})

		Convey("When everything is OK (vimeo)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://vimeo.com/89856635")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "Video posted successfully")
			})
		})

		Convey("When everything is OK (youtube)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://www.youtube.com/watch?v=9bZkp7q19f0")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "Video posted successfully")
			})
		})
	})
}