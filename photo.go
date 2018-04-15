package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

type photo struct {
	ID        uint
	UserID    uint
	Filename  string
	Caption   string
	CreatedAt time.Time
	Likes     uint
}

const thumbnailSize uint = 600

// FetchAllPhotos gets all photos for all users
func FetchAllPhotos(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)

	if uid == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := findUserByID(uid.(uint))

	if err != nil {
		log.Println("Could not find user:", err)
	}

	photos := []photo{}
	db.Order("id desc").Find(&photos)

	currentUser, _ := findUserByID(uid.(uint))

	c.HTML(http.StatusOK, "photos.html", gin.H{
		"user":        user,
		"photos":      photos,
		"CurrentUser": currentUser,
	})
}

// FetchSinglePhoto gets a single photo by ID
func FetchSinglePhoto(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)

	if uid == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Load single photo
	id := c.Params.ByName("id")
	photo := &photo{}
	db.Where("id = ?", id).Find(photo)

	// Load user info
	user, err := findUserByID(photo.UserID)

	if err != nil {
		log.Println("Could not find user:", err)
	}

	// Load comments
	comments := []comment{}
	db.Where("photo_id = ?", id).Find(&comments)

	currentUser, _ := findUserByID(uid.(uint))

	c.HTML(http.StatusOK, "photo.html", gin.H{
		"user":        user,
		"photo":       photo,
		"comments":    comments,
		"CurrentUser": currentUser,
	})
}

// CreatePhoto saves the file to disk, generates its thumbnails, and stores
// metadata in the database.
func CreatePhoto(c *gin.Context) {

	session := sessions.Default(c)
	uid := session.Get(userKey)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Could not find user: %s", uid))
		return
	}

	form, err := c.MultipartForm()

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	infile := form.File["photofile"][0]
	log.Println("Uploaded file:", infile.Filename)

	caption := form.Value["caption"][0]
	log.Println("Caption:", caption)

	uploadsdir := fmt.Sprintf("./public/uploads/%d", uid)

	if _, err := os.Stat(uploadsdir); os.IsNotExist(err) {
		os.Mkdir(uploadsdir, os.ModePerm)
	}

	thumbnailsdir := fmt.Sprintf("./public/thumbnails/%d", uid)

	if _, err := os.Stat(thumbnailsdir); os.IsNotExist(err) {
		os.Mkdir(thumbnailsdir, os.ModePerm)
	}

	// Generate unique filename

	ts := strconv.FormatInt(time.Now().UnixNano(), 10)
	fn := ts + filepath.Ext(infile.Filename)
	outfile := filepath.Join(uploadsdir, fn)

	// Save photo

	if err := c.SaveUploadedFile(infile, outfile); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	log.Println("Uploaded file:", outfile)

	// Insert DB record for photo and user

	photoid, err := insertPhoto(uid.(uint), fn, caption)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Insert photo err: %s", err.Error()))
		return
	}

	// Generate thumbnail

	err = generateThumbnail(uid.(uint), outfile, thumbnailSize)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error generating thumbnail: %s", err.Error()))
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/photos/%d", photoid))
}

// UpdatePhoto updates a single photo by ID
func UpdatePhoto(c *gin.Context) {

}

// DeletePhoto deletes a single photo by ID
func DeletePhoto(c *gin.Context) {
	id := c.Params.ByName("id")
	var p photo

	if err := db.Where("id = ?", id).Delete(&p).Error; err != nil {
		log.Println("Error deleting photo:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// LikePhoto increments the 'Likes' count
func LikePhoto(c *gin.Context) {
	photoid := c.Params.ByName("id")
	photo := &photo{}
	db.Where("id = ?", photoid).Find(photo)

	photo.Likes++

	if err := db.Save(photo); err.Error != nil {
		log.Println("Error updating photo:", err.Error)
	}

	c.JSON(http.StatusOK, gin.H{"likes": photo.Likes})
}

// CommentPhoto adds a comment to a photo
func CommentPhoto(c *gin.Context) {

	photoid, _ := strconv.ParseUint(c.Params.ByName("id"), 10, 64)

	var comment struct {
		Comment string `json:"comment"`
	}

	if err := c.BindJSON(&comment); err != nil {
		log.Println("BindJSON error:", err.Error())
	}

	log.Printf("Comment: %v\n", comment.Comment)

	session := sessions.Default(c)
	uid := session.Get(userKey)
	id, err := InsertComment(uint(photoid), uid.(uint), comment.Comment)

	if err != nil {
		log.Println("Error inserting comment:", err.Error())
	}

	user, _ := findUserByID(uid.(uint))

	c.JSON(http.StatusOK, gin.H{"id": id, "username": user.Username})
}

// Insert photo record into database
func insertPhoto(uid uint, fn string, caption string) (uint, error) {

	photo := &photo{
		UserID:    uid,
		Filename:  fn,
		Caption:   caption,
		CreatedAt: time.Now(),
	}

	if err := db.Create(photo); err.Error != nil {
		return 0, err.Error
	}

	log.Println("Inserted photo record:", photo.ID)

	return photo.ID, nil
}

func generateThumbnail(uid uint, photopath string, maxWidth uint) error {

	log.Println("Generating thumbnail for:", photopath)

	_, format, err := decodeConfig(photopath)

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Image format:", format)

	file, err := os.Open(photopath)
	if err != nil {
		log.Println("Error opening photo:", err)
	}

	var img image.Image

	switch format {
	case "jpeg":
		img, err = jpeg.Decode(file)
	case "png":
		img, err = png.Decode(file)
	case "gif":
		img, err = gif.Decode(file)
	default:
		err = errors.New("Unsupported file type")
	}

	if err != nil {
		log.Println("Error decoding photo:", err)
	}
	file.Close()

	log.Printf("Resizing image to %dpx\n", maxWidth)
	thumb := resize.Resize(maxWidth, 0, img, resize.Lanczos3)

	thumbnailPath := strings.Replace(photopath, "uploads", "thumbnails", -1)

	out, err := os.Create(thumbnailPath)

	if err != nil {
		log.Println("Error creating thumbnail path:", err)
	}

	defer out.Close()

	switch format {
	case "jpeg":
		err = jpeg.Encode(out, thumb, nil)
	case "png":
		err = png.Encode(out, thumb)
	case "gif":
		err = gif.Encode(out, thumb, nil)
	default:
		err = errors.New("Unsupported file type")
	}

	if err != nil {
		log.Println("Error encoding thumbnail:", err)
		return err
	}

	return nil
}

// Detect image file format (i.e. jpeg, png, gif)
func decodeConfig(filename string) (image.Config, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return image.Config{}, "", err
	}
	defer f.Close()
	return image.DecodeConfig(bufio.NewReader(f))
}

func (p *photo) TimeAgo() string {
	return humanize.Time(p.CreatedAt)
}
