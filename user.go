package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type user struct {
	ID       uint
	Email    string
	Username string
	Password string
	FullName string
}

type follower struct {
	UserID     uint `gorm:"unique_index:idx_user_follower"`
	FollowerID uint `gorm:"unique_index:idx_user_follower"`
}

const userKey = "userid"

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
	password := []byte(c.PostForm("password"))
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
		if bcrypt.CompareHashAndPassword([]byte(u.Password), password) != nil {
			session.AddFlash("Bad password")
			session.Save()
			c.HTML(http.StatusOK, "login.html", gin.H{
				"flash": session.Flashes(),
				"user":  u,
			})
		} else {
			log.Println("Saving user in session:", u.Username)
			session.Set(userKey, u.ID)
			session.Save()
			u := session.Get(userKey)
			log.Println("Testing user in session:", u)
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

	password := []byte(c.PostForm("password"))
	hashedPassword, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

	user := &user{
		FullName: c.PostForm("fullName"),
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
		Password: string(hashedPassword),
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

	log.Println("Creating DB user:", user.Username)

	if err := db.Create(user); err.Error != nil {
		log.Println("Error:", err.Error)
		session.AddFlash(err.Error)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": session.Flashes(),
			"user":  user,
		})
	} else {
		log.Println("Saving user in session:", user.Username)
		session.Set(userKey, user.ID)
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
		log.Println("Error:", err.Error)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	photos := []photo{}

	if err := db.Where("user_id = ?", user.ID).Order("id desc").Find(&photos); err.Error != nil {
		log.Println("Error:", err.Error)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	session := sessions.Default(c)
	uid := session.Get(userKey)
	currentUser, _ := findUserByID(uid.(uint))

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

func findUserByID(id uint) (*user, error) {
	u := &user{}

	if err := db.Where("id = ?", id).First(&u); err.Error != nil {
		return nil, err.Error
	}

	if u.ID == 0 {
		return nil, errors.New("User not found")
	}

	return u, nil
}

func (u *user) PhotoCount() uint {

	photos := []photo{}
	var count uint

	if err := db.Where("user_id = ?", u.ID).Find(&photos).Count(&count); err.Error != nil {
		log.Println("Error:", err.Error)
	}

	return count
}

// Follow inserts a record into the followers table
func Follow(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)
	fid, _ := strconv.ParseUint(c.Params.ByName("id"), 10, 64)

	follower := &follower{
		UserID:     uint(fid),
		FollowerID: uid.(uint),
	}

	if err := db.Create(follower); err.Error != nil {
		log.Println("Error:", err.Error)
	}

	c.JSON(http.StatusOK, nil)
}

// Unfollow deletes a record from the followers table
func Unfollow(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)
	fid, _ := strconv.ParseUint(c.Params.ByName("id"), 10, 64)

	follower := &follower{
		UserID:     uint(fid),
		FollowerID: uid.(uint),
	}

	if err := db.Where(&follower).Delete(follower); err.Error != nil {
		log.Println("Error:", err.Error)
	}

	c.JSON(http.StatusOK, nil)
}

func (u *user) Followers() uint {

	followers := []follower{}
	var count uint

	if err := db.Where("user_id = ?", u.ID).Find(&followers).Count(&count); err.Error != nil {
		log.Println("Error:", err.Error)
	}

	return count
}

func (u *user) Following() uint {

	followers := []follower{}
	var count uint

	if err := db.Where("follower_id = ?", u.ID).Find(&followers).Count(&count); err.Error != nil {
		log.Println("Error:", err.Error)
	}

	return count
}

// Follows returns true if the user (u) follows the userid
func (u *user) Follows(userid uint) bool {

	follower := &follower{
		UserID:     userid,
		FollowerID: u.ID,
	}

	if err := db.Where(&follower).Find(&follower); err.Error != nil {
		log.Println("Error:", err.Error)
	}

	return true
}
