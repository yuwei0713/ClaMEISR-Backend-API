package models

import (
	Routines "ginapi/routine"
	"ginapi/types"
)

func SearchUserData(Permission string) (any, any) {
	var db = Routines.MeisrDB
	var user = make([]types.FrontendUsers, 0)
	db.Table("users, userdatatable, schooltable").Select("DISTINCT userdatatable.* ,users.created_at ,schooltable.SchoolName").Where("users.username COLLATE utf8mb4_general_ci = userdatatable.Username").Where("userdatatable.SchoolCode = schooltable.SchoolCode").Find(&user)
	return user, nil
}
