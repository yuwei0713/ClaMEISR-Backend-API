package controllers

import (
	"fmt"
	"ginapi/models"
	"ginapi/types"
	"ginapi/utils"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var user types.BackendUsers
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	noneuse := models.RegisterVerify(user.Account)
	if noneuse { //可以使用該帳號名稱
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		// fmt.Println(string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = string(hashedPassword)
		fmt.Println(user)
		models.RegisterAccount(user)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

func Login(c *gin.Context) {
	var user types.BackendUsers
	session := sessions.Default(c)
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storedPassword, exists := models.LoginVerify(user.Account)
	if !exists || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		fmt.Print("Invalid username or password")
		return
	}

	token, err := utils.GenerateJWT(user.Account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	user = models.Login(user.Account)
	session.Set("profile", user)
	session.Save()
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"name":  user.Username, // Include the user profile data here
	})
}

func ClaMEISR_Register(c *gin.Context) {
	var user types.FrontendUsersRegister
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	fmt.Print(user)
	models.RegisterFrontEndAccount(user)

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}
