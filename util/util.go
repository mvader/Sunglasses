package util

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
	"unicode/utf8"
)

func Crypt(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 10)
	if err != nil {
		return "", err
	}

	return string(bytes[:]), nil
}

// RandomString returns a new random string
func RandomString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#$%&/()=?*+-_"
	var bytes = make([]byte, n)

	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}

	return string(bytes)
}

// NewRandomHash returns a random hash
func NewRandomHash() string {
	s := RandomString(25) + fmt.Sprint(time.Now().UnixNano())
	hasher := sha512.New()
	hasher.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// NewFileName returns an unique name for the file
func NewFileName(extension string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9]")
	return r.ReplaceAllString(NewRandomHash(), "") + "." + extension
}

// Hash hashes a string
func Hash(h string) string {
	hasher := sha512.New()
	hasher.Write([]byte(h))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// HashMD5 returns the md5 hash of a string
func HashMD5(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Strlen(s string) int {
	return utf8.RuneCountInString(s)
}

func IsValidURL(URL string) bool {
	r := regexp.MustCompile("https?://[-A-Za-z0-9+&@]+\\.[a-zA-Z0-9\\.-]+([/#\\?&\\.-_a-zA-Z0-9%=,:;$\\(\\)]+)?")
	return r.MatchString(URL)
}

func ResponseTitle(resp *http.Response) string {
	var title string

	r := regexp.MustCompile("<title>(.*)</title>")
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	matches := r.FindStringSubmatch(string(bytes))
	if len(matches) > 1 {
		title = matches[1]
	} else {
		title = "Untitled"
	}

	return title
}

func IsValidLink(URL string) (bool, string, string) {
	resp, err := http.Get(URL)
	if err != nil || resp.StatusCode != 200 {
		return false, "", ""
	}

	title := ResponseTitle(resp)

	return true, URL, title
}
