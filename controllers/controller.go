package controllers

import (
	"ginapi/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ShowQuestionnaire() gin.HandlerFunc {
	return func(c *gin.Context) {
		result := models.QuestionManage()
		c.JSON(200, gin.H{"Questionnaires": result})
	}
}
func ShowDetailQuestionnaire() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody map[string]map[string]string
		c.BindJSON(&requestBody)
		Keyvalue, _ := strconv.Atoi(requestBody["Param"]["Keyvalue"])
		result := models.DetailQuestionManage(Keyvalue)
		c.JSON(200, gin.H{"SearchData": result})
	}
}
func ShowSearch(SearchTitle string, Permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//SearchTitle 查找分類 Permission 權限
		result := models.SearchData(SearchTitle, Permission)
		c.JSON(200, gin.H{"SearchData": result})
	}
}
func ShowDetail(SearchTitle string, Permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody map[string]map[string]string
		c.BindJSON(&requestBody)
		Keyvalue := requestBody["Param"]["Keyvalue"]
		Basic, Detail := models.DetailSearch(SearchTitle, Permission, Keyvalue)
		c.JSON(200, gin.H{"BasicData": Basic, "DetailData": Detail})
	}
}
func UserManage(Motivation string, Permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var UserData, AdditionalText any
		// var Message string
		if Motivation == "Search" {
			UserData, AdditionalText = models.SearchUserData(Permission)
		} else {
			var requestBody map[string]map[string]string
			c.BindJSON(&requestBody)
		}
		c.JSON(200, gin.H{"UserData": UserData, "DetailData": AdditionalText})
	}
}
func SchoolManage(Motivation string, Permission string) gin.HandlerFunc {
	return func(c *gin.Context) {

		var SchoolData, AdditionalText any
		// var Message string
		if Motivation == "Search" {
			SchoolData, AdditionalText = models.SearchSchoolData(Permission)
		} else {
			var requestBody map[string]map[string]string
			c.BindJSON(&requestBody)
			if Motivation == "Update" {
				models.UpdateSchool(requestBody, Permission)
			}
			if Motivation == "Insert" {
				models.InsertSchool(requestBody, Permission)
			}
			if Motivation == "Delete" {

			}
		}
		c.JSON(200, gin.H{"SchoolData": SchoolData, "DetailData": AdditionalText})
	}
}
