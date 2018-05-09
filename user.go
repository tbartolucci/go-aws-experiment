package main

import (
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type user struct {
	ID       string `sql:"type:varchar(36);primary key"` // Cognito UUID
	Email    string
	Username string
	FullName string
}

type follower struct {
	UserID     string `gorm:"unique_index:idx_user_follower"`
	FollowerID string `gorm:"unique_index:idx_user_follower"`
}

const userKey = "userid"
const accessToken = "accessToken"

func loginForm(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login", "flash": flashes,
	})
}

func login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	u := &user{}
	session := sessions.Default(c)

	if err := db.Where("username = ?", username).First(&u); err.Error != nil {
		if err.RecordNotFound() {
			session.AddFlash("User not found")
		} else {
			session.AddFlash(err.Error)
		}

		session.Save()
		c.HTML(http.StatusOK, "login.html", gin.H{
			"flash": session.Flashes(),
			"user":  u,
		})
	} else {
		log.Info("Authenticating via Cognito: ", username)
		cog := NewCognito()
		jwt, err := cog.SignIn(username, password)

		if err != nil {
			msg := err.(awserr.Error).Message()
			log.Error("Signin Error: ", msg)
			session.AddFlash(msg)
			session.Save()
			c.HTML(http.StatusOK, "login.html", gin.H{
				"flash": session.Flashes(),
				"user":  u,
			})
		} else {
			log.Info("Authentication successful")
			sub, _ := cog.ValidateToken(jwt)
			session.Set(accessToken, jwt)
			session.Set(userKey, sub)
			session.Save()
			t := session.Get(accessToken)
			log.Info("Testing user in session:", t)
			c.Redirect(http.StatusFound, "/photos")
		}
	}
}

func signupForm(c *gin.Context) {
	session := sessions.Default(c)
	flashes := session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "signup.html", gin.H{"flash": flashes})
}

func signup(c *gin.Context) {
	user := &user{
		FullName: c.PostForm("fullName"),
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
	}

	session := sessions.Default(c)

	u, _ := findUserByUsername(user.Username)

	if u != nil {
		msg := "This username isn't available. Please try another."
		session.AddFlash(msg)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": session.Flashes(),
			"user":  user,
		})
		session.Save()
		return
	}

	cog := NewCognito()
	password := c.PostForm("password")
	jwt, err := cog.SignUp(user.Username, password, user.Email, user.FullName)

	if err != nil {
		msg := err.(awserr.Error).Message()
		log.Error("SignUp error: ", msg)
		session.AddFlash(msg)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": session.Flashes(),
			"user":  user,
		})
		session.Save()
		return
	}

	log.Info("Creating DB user:", user.Username)

	sub, err := cog.ValidateToken(jwt)

	if err != nil {
		return
	}

	log.Info("Cognito 'sub': ", sub)

	user.ID = sub // Set user ID to Cognito UUID

	if err := db.Create(user); err.Error != nil {
		log.Error("Error:", err.Error)
		session.AddFlash(err.Error)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": session.Flashes(),
			"user":  user,
		})
	} else {
		log.Info("Saving userid in session for: ", user.Username)
		session.Set(userKey, user.ID)
		session.Set(accessToken, jwt)
		session.Save()
		c.Redirect(http.StatusFound, "/photos")
	}

	session.Save()
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	c.Redirect(302, "/")
}

// Profile shows the user profile
// GET /user/:username
func Profile(c *gin.Context) {
	user := &user{Username: c.Params.ByName("username")}

	if err := db.Where(&user).First(&user); err.Error != nil {
		log.Error("Error:", err.Error)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	photos := []photo{}

	if err := db.Where("user_id = ?", user.ID).Order("id desc").Find(&photos); err.Error != nil {
		log.Error("Error:", err.Error)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	session := sessions.Default(c)
	uid := session.Get(userKey)
	currentUser, _ := findUserByID(uid.(string))

	c.HTML(http.StatusOK, "user.html", gin.H{
		"user":        user,
		"photos":      photos,
		"IsSelf":      uid == user.ID,
		"CurrentUser": currentUser,
	})
}

func findUserByUsername(username string) (*user, error) {
	u := &user{}

	if err := db.Where("username = ?", username).First(&u); err.Error != nil {
		return nil, err.Error
	}

	return u, nil
}

func findUserByID(id string) (*user, error) {
	u := &user{}

	if err := db.Where("id = ?", id).First(&u); err.Error != nil {
		return nil, err.Error
	}

	if u.ID == "" {
		return nil, errors.New("User not found")
	}

	return u, nil
}

func (u *user) PhotoCount() uint {

	photos := []photo{}
	var count uint

	if err := db.Where("user_id = ?", u.ID).Find(&photos).Count(&count); err.Error != nil {
		log.Error("Error:", err.Error)
	}

	return count
}

// Follow inserts a record into the followers table
func Follow(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)
	fid := c.Params.ByName("id")

	follower := &follower{
		UserID:     fid,
		FollowerID: uid.(string),
	}

	if err := db.Create(follower); err.Error != nil {
		log.Error("Error:", err.Error)
	}

	c.JSON(http.StatusOK, nil)
}

// Unfollow deletes a record from the followers table
func Unfollow(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)
	fid := c.Params.ByName("id")

	follower := &follower{
		UserID:     fid,
		FollowerID: uid.(string),
	}

	if err := db.Where(&follower).Delete(follower); err.Error != nil {
		log.Error("Error:", err.Error)
	}

	c.JSON(http.StatusOK, nil)
}

func (u *user) Followers() uint {

	followers := []follower{}
	var count uint

	if err := db.Where("user_id = ?", u.ID).Find(&followers).Count(&count); err.Error != nil {
		log.Error("Error:", err.Error)
	}

	return count
}

func (u *user) Following() uint {

	followers := []follower{}
	var count uint

	if err := db.Where("follower_id = ?", u.ID).Find(&followers).Count(&count); err.Error != nil {
		log.Error("Error:", err.Error)
	}

	return count
}

// Follows returns true if the user (u) follows the userid
func (u *user) Follows(userid string) bool {

	follower := &follower{
		UserID:     userid,
		FollowerID: u.ID,
	}

	if err := db.Where(&follower).Find(&follower); err.Error != nil {
		log.Error("Error:", err.Error)
	}

	return true
}
