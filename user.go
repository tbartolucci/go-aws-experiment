package main

import (
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
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

	if _, err := findUserByUsername(username); err != nil {
		session.AddFlash(err.Error)

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

	sessionStore := sessions.Default(c)

	u, _ := findUserByUsername(user.Username)

	if u != nil {
		msg := "This username isn't available. Please try another."
		sessionStore.AddFlash(msg)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})
		sessionStore.Save()
		return
	}

	cog := NewCognito()
	password := c.PostForm("password")
	jwt, err := cog.SignUp(user.Username, password, user.Email, user.FullName)

	if err != nil {
		msg := err.(awserr.Error).Message()
		log.Error("SignUp error: ", msg)
		sessionStore.AddFlash(msg)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})
		sessionStore.Save()
		return
	}

	log.Info("Creating DB user:", user.Username)

	sub, err := cog.ValidateToken(jwt)

	if err != nil {
		return
	}

	log.Info("Cognito 'sub': ", sub)

	user.ID = sub // Set user ID to Cognito UUID

	av, err := dynamodbattribute.MarshalMap(user)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
		sessionStore.AddFlash(err)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})

		c.Redirect(http.StatusFound, "/photos")
		sessionStore.Save()
		return
	}

	svc := NewDynamoDb()

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("PhotoAppUsers"),
		Item: av,
	})

	if err != nil {
		log.Errorf("Error: %v", err)
		sessionStore.AddFlash(err)
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"flash": sessionStore.Flashes(),
			"user":  user,
		})
	} else {
		log.Info("Saving userid in session for: ", user.Username)
		sessionStore.Set(userKey, user.ID)
		sessionStore.Set(accessToken, jwt)
		sessionStore.Save()
		c.Redirect(http.StatusFound, "/photos")
	}

	sessionStore.Save()
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
	username := c.Params.ByName("username")
	user, err := findUserByUsername(username)

	if err != nil {
		log.Error("Error:", err.Error)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppPhotos"),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID" : {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(user.ID),
					},
				},
			},
		},
		IndexName: aws.String("UserID-index"),
	}

	svc := NewDynamoDb()
	qo, err := svc.Query(queryInput)
	if err != nil {
		log.Errorf("Error: %v", err)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	photos := []photo{}
	// deserialize items into photos list variable
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &photos)
	if err != nil {
		log.Errorf("failed to unmarshal Query result items, %v", err)
		c.HTML(http.StatusOK, "404.html", nil)
		return
	}

	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)
	currentUser, _ := findUserByID(uid.(string))

	c.HTML(http.StatusOK, "user.html", gin.H{
		"user":        user,
		"photos":      photos,
		"IsSelf":      uid == user.ID,
		"CurrentUser": currentUser,
	})
}

func findUserByUsername(username string) (*user, error) {

	sess := NewAwsSession()
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppUsers"),
		Limit: aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"Username": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(username),
					},
				},
			},
		},
		IndexName: aws.String("Username-index"),
	}

	qo, err := svc.Query(queryInput)

	return findUserHelper(qo, err)
}

func findUserByID(id string) (*user, error) {
	sess := NewAwsSession()
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppUsers"),
		Limit: aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"ID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(id),
					},
				},
			},
		},
		IndexName: aws.String("Username-index"),
	}

	qo, err := svc.Query(queryInput)

	return findUserHelper(qo, err)
	return findUserHelper(qo, err)
}

func findUserHelper(qo *dynamodb.QueryOutput, err error) (*user, error) {
	if err != nil {
		log.Errorf("FindUserByUsername failed: %v", err)
		return nil, err
	}

	users := []user{}
	if err := dynamodbattribute.UnmarshalListOfMaps(qo.Items, &users); err != nil {
		log.Errorf("Failed to unmarshal Query result items, %v", err)
		return nil, err
	}

	if len(users) == 0 {
		// Returned no users
		return nil, errors.New("User not found")
	}

	return &users[0], nil
}

func (u *user) PhotoCount() uint {
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppPhotos"),
		Select: aws.String("COUNT"),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID" : {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
		},
		IndexName: aws.String("UserID-index"),
	}

	svc := NewDynamoDb()
	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error getting photo count: %v", err)
		return 0
	}

	count := aws.Int64Value(qo.Count)

	return uint(count)
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

	av, err := dynamodbattribute.MarshalMap(follower)
	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
	}

	svc := NewDynamoDb()

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("PhotosAppUsers"),
		Item: av,
	})

	if err != nil {
		log.Errorf("failed to put record to DynamoDB, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// Unfollow deletes a record from the followers table
func Unfollow(c *gin.Context) {
	sessionStore := sessions.Default(c)
	uid := sessionStore.Get(userKey)
	fid := c.Params.ByName("id")

	follower := &follower{
		UserID:     fid,
		FollowerID: uid.(string),
	}

	// Delete follower from DynamoDB

	av, err := dynamodbattribute.MarshalMap(follower)

	if err != nil {
		log.Errorf("failed to DynamoDB marshal Record, %v", err)
	}

	svc := NewDynamoDb()

	_, err = svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("PhotosAppUsers"),
		Key:       av,
	})

	if err != nil {
		log.Errorf("failed to delete record from DynamoDB, %v", err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	c.JSON(http.StatusOK, nil)
}

func (u *user) Followers() uint {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppFollowers"),
		Select:    aws.String("COUNT"),
		KeyConditions: map[string]*dynamodb.Condition{
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
		},
	}

	svc := NewDynamoDb()

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error getting follower count: %v", err)
		return 0
	}

	count := aws.Int64Value(qo.Count)

	return uint(count)
}

func (u *user) Following() uint {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppFollowers"),
		Select:    aws.String("COUNT"),
		KeyConditions: map[string]*dynamodb.Condition{
			"FollowerID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
		},
		IndexName: aws.String("FollowerID-index"),
	}

	svc := NewDynamoDb()

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error getting following count: %v", err)
		return 0
	}

	count := aws.Int64Value(qo.Count)

	return uint(count)
}

// Follows returns true if the user (u) follows the userid
func (u *user) Follows(userid string) bool {

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("PhotosAppFollowers"),
		Select:    aws.String("COUNT"),

		KeyConditions: map[string]*dynamodb.Condition{
			"FollowerID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(u.ID),
					},
				},
			},
			"UserID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userid),
					},
				},
			},
		},
	}

	svc := NewDynamoDb()

	qo, err := svc.Query(queryInput)

	if err != nil {
		log.Errorf("Error getting follows count: %v", err)
		return false
	}

	count := aws.Int64Value(qo.Count)

	return count > 0
}
