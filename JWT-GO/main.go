package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type users struct {
	UID      uint   `gorm:"primaryKey;autoIncrement"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

var Db *gorm.DB
var SQL *sql.DB

func Initdb() gin.HandlerFunc {
	var err error
	dsn := "root:==bitstek@700@tcp(127.0.0.1:3306)/users?charset=utf8mb4&parseTime=True&loc=Local"
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to db ", err)
	}
	return func(c *gin.Context) {
		// Db.AutoMigrate(&users{})
		SQL, err := Db.DB()
		if err != nil || SQL.Ping() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database connection failed",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func createToken() {

}

// -----------<<<<< main >>>>>----------------
func main() {
	router := gin.Default()

	router.Use(Initdb())
	router.GET("/login/:email", login)
	router.POST("/signup", signup)
	router.Run(":700")
	defer SQL.Close()
}

//-----------<<<<< main end >>>>>----------------

func login(c *gin.Context) {
	email := c.Param("email")
	var input users
	var cred users
	err := c.ShouldBindJSON(&input) // json into struct
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		inputPassword := input.Password
		err = Db.Where("email =?", email).First(&cred).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "user doesnt exist please signup !"})
			c.Redirect(http.StatusNotModified, "/signup")
			return
		}
		// ComparePasswords compares the hashed password with the plain password
		if err = bcrypt.CompareHashAndPassword([]byte(cred.Password), []byte(inputPassword)); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "password dosent match !!!",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "login sucessful",
		})
		//	c.Redirect(http.StatusAccepted, "/welcome.html")
	}()
	wg.Wait()
	//defer SQL.Close() // this is causing race condition .. sql sever nil pointer error
}
func signup(c *gin.Context) {
	var cred users
	err := json.NewDecoder(c.Request.Body).Decode(&cred) // decoding json into struct
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost) // storing the hashed password
		if err != nil {                                                                               // can set 14 ,16 ,20 ...
			return
		}
		cred.Password = string(hashedPassword)
		if err = Db.First(&cred, "email =?", cred.Email).Error; err != nil { //first checking does user already exist??

			if err == gorm.ErrRecordNotFound { //if doesnt exist error came ... we create the new user
				if erro := Db.Create(&cred).Error; erro != nil { // creating a user in db
					c.JSON(http.StatusInternalServerError, gin.H{"error ": erro.Error()})
					return
				}
				c.JSON(http.StatusAccepted, gin.H{
					"a message ": "user created try to login now",
				})
				return
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else { // no error from first if condition then... it means it exist
			c.JSON(http.StatusOK, gin.H{
				"a message": "this email already exist!! :",
				"details":   cred.Email,
				"message":   "try to login!! :",
			})
			return
		}
	}()
	wg.Wait()
	//defer SQL.Close() // this is causing race condition .. sql sever nil pointer error
}
