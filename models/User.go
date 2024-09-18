package models

import (
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"
	"time"
)

func RegisterVerify(Account string) bool { //帳號是否重複
	var db = Routines.BackStageDB
	var name string
	db.Table("users").Select("username").Where("username = ?", Account).Take(&name)

	if name == "" {
		return true
	} else {
		return false
	}
}

func RegisterAccount(User types.BackendUsers) { //建立帳號
	var db = Routines.BackStageDB
	fmt.Println(User)
	currentTime := time.Now()
	formattime := currentTime.Format("2006-01-02 15:04:05")
	result := db.Exec("INSERT INTO users (account, username, schoolnumber, password, permission, created_at, updated_at) VALUES (?,?,?,?,?,?,?)", User.Account, User.Username, User.SchoolCode, User.Password, User.Permission, formattime, formattime)
	fmt.Println(result)
}

func LoginVerify(Account string) (string, bool) { //驗證帳號
	//storedPassword, exists
	var db = Routines.BackStageDB
	var exists string
	VarifyPassword := ""
	db.Table("users").Select("account").Where("account = ?", Account).Take(&exists)

	if exists != "" {
		db.Table("users").Select("password").Where("account = ?", Account).Take(&VarifyPassword)
		return VarifyPassword, true
	} else {
		return VarifyPassword, false
	}
}

func Login(Account string) types.BackendUsers { //獲取帳號資料
	var db = Routines.BackStageDB
	var UserData types.BackendUsers
	result := db.Table("users").Select("*").Where("account = ?", Account).Take(&UserData)

	// Check if no records were found
	if result.RowsAffected == 0 {
		// No record found
		fmt.Println("No record found.")
	} else if result.Error != nil {
		// An error occurred during the query
		fmt.Println("Error occurred:", result.Error)
	}
	return UserData
}

func Logout() { //登出

}
