package controllers

import (
	"fmt"
	"ginapi/models"
	"net/http"
	"strconv"
	"strings"

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
		Basic, Detail, School := models.DetailSearch(SearchTitle, Permission, Keyvalue)
		c.JSON(200, gin.H{"BasicData": Basic, "DetailData": Detail, "SchoolData": School})
	}
}
func ChildManage(Motivation string, Permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if Motivation == "Update" {
			var FormData map[string]interface{}
			c.BindJSON(&FormData)
			result := models.UpdateChildInfo(FormData)
			c.JSON(200, gin.H{"status": result})
		}
	}
}
func UserManage(Motivation string, Permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var UserData, AdditionalText any
		// var Message string
		if Motivation == "Search" {
			UserData, AdditionalText = models.SearchUserData(Permission)
		} else if Motivation == "Update" {
			var FormData map[string]interface{}
			c.BindJSON(&FormData)
			result := models.UpdateUserData(FormData, Permission)
			c.JSON(200, gin.H{"status": result})
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
				models.DeleteSchool(requestBody, Permission)
			}
		}
		c.JSON(200, gin.H{"SchoolData": SchoolData, "DetailData": AdditionalText})
	}
}
func ExportToExcel() gin.HandlerFunc {
	return func(c *gin.Context) {
		yearSemester := c.Query("year_semester")

		// 檢查參數是否存在
		if yearSemester == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing year_semester parameter"})
			return
		}
		// 將 year_semester 分成 Year 和 Semester
		split := strings.Split(yearSemester, "-")
		if len(split) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year_semester format"})
			return
		}

		Year := strings.TrimSpace(split[0])
		Semester := strings.TrimSpace(split[1])
		// 調用模型函數產生Excel檔案
		excelFile := models.DefaultExport(Year, Semester)
		// 設置回應標頭來使瀏覽器下載Excel檔案
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment;filename=sample.xlsx")
		c.Header("File-Name", fmt.Sprintf("ClaMEISR 資料分析表_%s-%s.xlsx", Year, Semester))
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Expires", "0")

		// 將Excel寫入HTTP回應
		if err := excelFile.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write Excel file"})
			return
		}
	}
}
