package models

import (
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"

	"gorm.io/gorm"
)

func SearchUserData(Permission string) (any, any) {
	var db = Routines.MeisrDB
	var user = make([]types.FrontendUsers, 0)
	db.Table("users, userdatatable, schooltable").Select("DISTINCT userdatatable.* ,users.created_at ,schooltable.SchoolName").Where("users.username COLLATE utf8mb4_general_ci = userdatatable.Username").Where("userdatatable.SchoolCode = schooltable.SchoolCode").Find(&user)
	return user, nil
}

func UpdateUserData(UserData map[string]interface{}, Permission string) bool {
	var db = Routines.MeisrDB
	var user types.FrontendUsers
	paramData, _ := UserData["Param"].(map[string]interface{})

	teacherData, _ := paramData["TeacherData"].(map[string]interface{})
	account := teacherData["Account"].(string)
	counseling := teacherData["Counseling"].(string)
	kindergarten := teacherData["Kindergarten"].(string)
	routinesBased := teacherData["RoutinesBased"].(string)
	schoolCodeFloat64, _ := teacherData["SchoolCode"].(float64)
	schoolCode := int(schoolCodeFloat64)
	sperate := teacherData["Sperate"].(string)
	teacherName := teacherData["TeacherName"].(string)

	result := db.Table("userdatatable").Select("Username").Where("Username = ? AND SchoolCode = ?", account, schoolCode).Find(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("User not found")
		} else {
			fmt.Printf("Error occurred: %v\n", result.Error)
		}
	} else {
		db.Exec("UPDATE userdatatable SET TeacherName = ?, Separate = ?, Kindergarten = ?, Counseling = ?, RoutinesBased = ?  WHERE Username = ? AND SchoolCode = ?", teacherName, sperate, kindergarten, counseling, routinesBased, account, schoolCode)
		return true
	}
	return false
}
