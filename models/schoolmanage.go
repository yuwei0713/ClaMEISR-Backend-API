package models

import (
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"
	"strconv"
)

func SearchSchoolData(Permission string) (any, any) {
	var db = Routines.MeisrDB
	var TotalSchool = make([]types.Schools, 0)
	db.Table("schooltable").Select("DISTINCT SchoolName ,SchoolCode").Find(&TotalSchool)
	for SchoolCount := 0; SchoolCount < len(TotalSchool); SchoolCount++ {
		var EachClass = make([]types.Classes, 0)
		db.Table("schooltable").Select("ClassCode, ClassName").Where("SchoolCode", TotalSchool[SchoolCount].SchoolCode).Find(&EachClass)
		TotalSchool[SchoolCount].Classes = EachClass
	}
	return TotalSchool, nil
}
func InsertSchool(DealData map[string]map[string]string, Permission string) {
	var db = Routines.MeisrDB
	Target := DealData["Param"]["Target"]
	if Target == "Class" {
		var SchoolName = ""
		SchoolCode, _ := strconv.Atoi(DealData["Param"]["SchoolCode"])
		ClassCode, _ := strconv.Atoi(DealData["Param"]["ClassCode"])
		ClassName := DealData["Param"]["ClassName"]
		db.Table("schooltable").Select("DISTINCT SchoolName").Where("SchoolCode", SchoolCode).Find(&SchoolName)
		result := db.Exec("INSERT INTO schooltable (SchoolName,SchoolCode,ClassName,ClassCode) VALUES (?,?,?,?)", SchoolName, SchoolCode, ClassName, ClassCode)
		fmt.Print(result)
	}
	if Target == "School" {
		fmt.Println("Get in!")
		SchoolName := DealData["Param"]["SchoolName"]
		fmt.Println("Get SchoolName!", SchoolName)
		SchoolCode, _ := strconv.Atoi(DealData["Param"]["SchoolCode"])
		fmt.Println("Get SchoolCode!", SchoolCode)
		DefaultClassName := DealData["Param"]["ClassName"]
		fmt.Println("Get ClassName!", DefaultClassName)
		ClassesNumber, _ := strconv.Atoi(DealData["Param"]["ClassesNumber"])
		fmt.Println("Get Classes Number!", ClassesNumber)

		for ClassCode := 1; ClassCode <= ClassesNumber; ClassCode++ {
			ClassName := string(DefaultClassName + strconv.Itoa(ClassCode))
			result := db.Exec("INSERT INTO schooltable (SchoolName,SchoolCode,ClassName,ClassCode) VALUES (?,?,?,?)", SchoolName, SchoolCode, ClassName, ClassCode)
			fmt.Print(result)
		}
	}
}
func UpdateSchool(DealData map[string]map[string]string, Permission string) {
	var db = Routines.MeisrDB
	Target := DealData["Param"]["Target"]
	if Target == "Class" {
		SchoolCode, _ := strconv.Atoi(DealData["Param"]["SchoolCode"])
		ClassCode, _ := strconv.Atoi(DealData["Param"]["ClassCode"])
		ClassName := DealData["Param"]["ClassName"]
		result := db.Exec("UPDATE schooltable SET ClassName = ? WHERE SchoolCode = ? AND ClassCode = ?", ClassName, SchoolCode, ClassCode)
		fmt.Print(result)
	}
}
func DeleteSchool(DealData map[string]map[string]string, Permission string) {
	var db = Routines.MeisrDB
	Target := DealData["Param"]["Target"]
	if Target == "Class" {
		SchoolCode, _ := strconv.Atoi(DealData["Param"]["SchoolCode"])
		ClassCode, _ := strconv.Atoi(DealData["Param"]["ClassCode"])
		result := db.Exec("DELETE from schooltable WHERE SchoolCode = ? AND ClassCode = ?", SchoolCode, ClassCode)
		fmt.Print(result)
	}
	if Target == "School" {
		SchoolCode, _ := strconv.Atoi(DealData["Param"]["SchoolCode"])
		result := db.Exec("DELETE from schooltable WHERE SchoolCode = ?", SchoolCode)
		fmt.Print(result)
	}
}
