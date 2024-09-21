package models

import (
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"
	"ginapi/utils"
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
func RegisterFrondendVerify(Account string) bool { //帳號是否重複
	var db = Routines.MeisrDB
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

func RegisterFrontEndAccount(User types.FrontendUsers_Register) bool {
	var db = Routines.MeisrDB
	if User.Quantity != 0 { //批量新增
		createvalue := 1
		currentnumber := 1
		currentTime := time.Now()
		formattime := currentTime.Format("2006-01-02 15:04:05")
		for {
			CurrentAccount := fmt.Sprintf("%s%d", User.Account, currentnumber)
			CurrentPassword := fmt.Sprintf("%s%d", User.Password, currentnumber)
			noneuse := RegisterFrondendVerify(CurrentAccount)
			if noneuse { //can create
				hashedPassword, err := utils.BcryptHash(CurrentPassword) //建立密碼
				if err != nil {
					// c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
					return false
				}
				db.Exec("INSERT INTO users (username, schoolnumber, password, created_at, updated_at) VALUES(?,?,?,?,?)", User.Account, User.SchoolCode, hashedPassword, formattime, formattime)
				db.Exec("INSERT INTO userdatatable (Username, SchoolCode) VALUES (?,?)", User.Account, User.SchoolCode)
				createvalue++
			}
			currentnumber++
			if createvalue > User.Quantity {
				break
			}
		}
		return true
	} else { //單一新增
		noneuse := RegisterFrondendVerify(User.Account)
		if noneuse { //可以使用該帳號名稱
			hashedPassword, err := utils.BcryptHash(User.Password) //建立密碼
			// fmt.Println(string(hashedPassword))
			if err != nil {
				// c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
				return false
			}
			currentTime := time.Now()
			formattime := currentTime.Format("2006-01-02 15:04:05")
			db.Exec("INSERT INTO users (username, schoolnumber, password, created_at, updated_at) VALUES(?,?,?,?,?)", User.Account, User.SchoolCode, hashedPassword, formattime, formattime)
			db.Exec("INSERT INTO userdatatable (Username, SchoolCode) VALUES (?,?)", User.Account, User.SchoolCode)
			return true
		}
	}

	return true
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
