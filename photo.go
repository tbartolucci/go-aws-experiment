package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type photo struct {
	ID        uint
	UserID    string `sql:"type:varchar(36);primary key"` // Cognito UUID
	Filename  string
	Caption   string
	CreatedAt time.Time
	Likes     uint
}

const thumbnailSize uint = 600

var bucketName string

func init() {

	log.Info("Initializing S3")

	log.Info("Loading configuration")
	viper.SetConfigName("config") // config.toml
	viper.AddConfigPath(".")      // use working directory

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("error reading config file, %v", err)
		return
	}

	bucketName = viper.GetString("s3.bucketName")

	log.Info("S3 bucket: ", bucketName)
}

// FetchAllPhotos gets all photos for all users
func FetchAllPhotos(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get(userKey)

	if uid == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := findUserByID(uid.(string))

	if err != nil {
		log.Error("Could not find user:", err)
	}

	photos := []photo{}
	db.Order("id desc").Find(&photos)

	currentUser, _ := findUserByID(uid.(string))

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
		log.Error("Could not find user:", err)
	}

	// Load comments
	comments := []comment{}
	db.Where("photo_id = ?", id).Find(&comments)

	currentUser, _ := findUserByID(uid.(string))

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
	jwt := session.Get(accessToken)
	cog := NewCognito()
	sub, _ := cog.ValidateToken(jwt.(string))

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Could not find user: %s", sub))
		return
	}

	form, err := c.MultipartForm()

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	file, header, err := c.Request.FormFile("photofile")

	if err != nil {
		log.Errorf("Error uploading file %v", err)
		return
	}

	defer file.Close()

	caption := form.Value["caption"][0]
	log.Info("Caption:", caption)

	// Upload file to S3 bucket

	sess := awsSession.Must(awsSession.NewSession())
	uploader := s3manager.NewUploader(sess)

	key := sub + "/" + header.Filename

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(mime.TypeByExtension(filepath.Ext(header.Filename))),
	})

	if err != nil {
		log.Errorf("Unable to upload file %q, %v", header.Filename, err)
		c.String(http.StatusBadRequest, fmt.Sprintf("Upload file err: %s", err.Error()))
		return
	}

	log.Info("Uploaded file:", header.Filename)

	// Insert DB record for photo and user

	photoid, err := insertPhoto(sub, header.Filename, caption)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Insert photo err: %s", err.Error()))
		return
	}

	// Generate thumbnail

	err = generateThumbnail(sess, sub, header.Filename, key, thumbnailSize)

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
		log.Error("Error deleting photo:", err)
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
		log.Error("Error updating photo:", err.Error)
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
		log.Error("BindJSON error:", err.Error())
	}

	log.Printf("Comment: %v\n", comment.Comment)

	session := sessions.Default(c)
	uid := session.Get(userKey)
	id, err := InsertComment(uint(photoid), uid.(string), comment.Comment)

	if err != nil {
		log.Error("Error inserting comment:", err.Error())
	}

	user, _ := findUserByID(uid.(string))

	c.JSON(http.StatusOK, gin.H{"id": id, "username": user.Username})
}

// Insert photo record into database
func insertPhoto(uid string, fn string, caption string) (uint, error) {

	photo := &photo{
		UserID:    uid,
		Filename:  fn,
		Caption:   caption,
		CreatedAt: time.Now(),
	}

	if err := db.Create(photo); err.Error != nil {
		return 0, err.Error
	}

	log.Info("Inserted photo record:", photo.ID)

	return photo.ID, nil
}

func generateThumbnail(sess *awsSession.Session, sub string, filename string, key string, maxWidth uint) error {

	log.Infof("Fetching s3://%v/%v", bucketName, key)

	buff := &aws.WriteAtBuffer{}
	s3dl := s3manager.NewDownloader(sess)
	_, err := s3dl.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Fatalf("Could not download from S3: %v", err)
	}

	log.Infof("Decoding image")

	imageBytes := buff.Bytes()
	reader := bytes.NewReader(imageBytes)

	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Fatalf("bad response: %s", err)
	}

	log.Infof("Generating thumbnail")
	thumbnail := resize.Thumbnail(maxWidth, maxWidth, img, resize.Lanczos3)

	log.Infof("Encoding image for upload to S3")
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumbnail, nil)

	if err != nil {
		log.Errorf("JPEG encoding error: %v", err)
		return err
	}

	thumbkey := sub + "/thumb/" + filename

	log.Infof("Preparing S3 object: %s", thumbkey)

	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:        bytes.NewReader(buf.Bytes()),
		Bucket:      aws.String(bucketName),
		Key:         aws.String(thumbkey),
		ContentType: aws.String(mime.TypeByExtension(filepath.Ext(filename))),
	})

	if err != nil {
		log.Error("Failed to upload", err)
		return err
	}

	log.Println("Successfully uploaded to", result.Location)

	return nil
}

func (p *photo) TimeAgo() string {
	return humanize.Time(p.CreatedAt)
}
