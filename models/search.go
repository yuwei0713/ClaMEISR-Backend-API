package models

import (
	Routines "ginapi/routine"
	"ginapi/types"
	"strings"
)

func SearchData(SearchType string, Permission string) any {
	var db = Routines.MeisrDB
	if SearchType == "Teacher" {
		var users = make([]types.SearchUsers, 0)
		db.Table("userdatatable, schooltable").Select("DISTINCT userdatatable.*, schooltable.SchoolName").Where("userdatatable.IfFill = ?", 1).Where("userdatatable.SchoolCode = schooltable.SchoolCode").Find(&users) //select * from 'userdatatable'
		return users
	} else if SearchType == "Children" {
		var BasicChild = make([]types.Child, 0)
		db.Table("studentschooltable as childbasic, userdatatable as user").Select("childbasic.* ,user.TeacherName as TeacherName").Where("childbasic.SchoolCode = user.SchoolCode").Where("childbasic.TeacherAccount = user.Username").Find(&BasicChild)
		return BasicChild
	} else if SearchType == "FillStatus" {
		var QuestionFill = make([]types.QuestionFill, 0)
		var Child = make([]types.Child, 0)
		db.Table("studentschooltable as childbasic, userdatatable as user").Select("childbasic.* ,user.TeacherName as TeacherName").Where("childbasic.SchoolCode = user.SchoolCode").Where("childbasic.TeacherAccount = user.Username").Find(&Child)
		for i := 0; i < len(Child); i++ {
			var FillStatus = make([]types.FillStatus, 0)
			db.Table("studentfillfinish").Select("SchoolYear as Year, Semester, FillTime, Finish").Where("StudentID = ?", Child[i].StudentID).Find(&FillStatus)
			for j := 0; j < len(FillStatus); j++ {
				var Time = FillStatus[j].FillTime
				var Finish = FillStatus[j].Finish
				var FillYear = FillStatus[j].Year
				var FillSemester = FillStatus[j].Semester
				var FillData types.FillData
				var CombineData types.QuestionFill
				if Finish == 1 {
					for FillTime := 1; FillTime <= Time; FillTime++ {
						db.Table("questionstoretable as filldata, questionnaireframework as questiondata").Select("filldata.SchoolYear as FillYear, filldata.Semester as FillSemester, filldata.FillTime, filldata.FillDate, questiondata.QuestionCode, questiondata.QuestionName").Where("filldata.FillTime = ?", FillTime).Where("filldata.SchoolYear = ?", FillYear).Where("filldata.Semester = ?", FillSemester).Where("filldata.QuestionCode = questiondata.QuestionCode").Take(&FillData)
						CombineData.Child = Child[i]
						CombineData.FillData = FillData
						QuestionFill = append(QuestionFill, CombineData)
					}
				} else if Finish == 0 {
					for FillTime := 1; FillTime < Time; FillTime++ {
						db.Table("questionstoretable as filldata, questionnaireframework as questiondata").Select("filldata.SchoolYear as FillYear, filldata.Semester as FillSemester, filldata.FillTime, filldata.FillDate, questiondata.QuestionCode, questiondata.QuestionName").Where("filldata.FillTime = ?", FillTime).Where("filldata.SchoolYear = ?", FillYear).Where("filldata.Semester = ?", FillSemester).Where("filldata.QuestionCode = questiondata.QuestionCode").Take(&FillData)
						CombineData.Child = Child[i]
						CombineData.FillData = FillData
						QuestionFill = append(QuestionFill, CombineData)
					}
				}
			}
		}
		// db.Table("studentschooltable as childbasic, studentfillfinish as fillbasic, questionstoretable as filldetail, questionnaireframework as questiontitle").Select("childbasic.*").Select("fillbasic.SchoolYear as FillYear, fillbasic.Semester as FillSemester").Where("fillbasic.StudentID = childbasic.StudentID").Select("questiontitle.QuestionName").Where("fillbasic.QuestionCode = questiontitle.QuestionCode").Select("filldetail.FillTime, filldetail.FillDate").Where("fillbasic.FillTime > 0").Where("filldetail.FillTime <= fillbasic.FillTime").Find(&QuestionFill)
		return QuestionFill
	} else {
		return false
	}
}
func DetailSearch(SearchType string, Permission string, Keyvalue string) (any, any) {
	var db = Routines.MeisrDB
	if SearchType == "Teacher" {
		var users types.SearchUsers
		db.Table("userdatatable, schooltable").Select("DISTINCT userdatatable.*, schooltable.SchoolName").Where("userdatatable.IfFill = ?", 1).Where("userdatatable.SchoolCode = schooltable.SchoolCode").Where("userdatatable.Username = ?", Keyvalue).Take(&users)
		var BasicChild = make([]types.Child, 0)
		db.Table("studentschooltable as childbasic").Select("childbasic.*").Where("childbasic.TeacherAccount = ?", Keyvalue).Find(&BasicChild)
		return users, BasicChild
	} else if SearchType == "Children" {
		var ChildAllData types.ChildDetail
		var Child types.Child
		var ChildDiagnosis types.ChildDiagnosis
		var ChildFamily types.ChildFamily

		var FillData = make([]types.FillData, 0)

		db.Table("studentschooltable as childbasic, userdatatable as user").Select("childbasic.* ,user.TeacherName as TeacherName").Where("childbasic.TeacherAccount = user.Username").Where("childbasic.StudentID = ?", Keyvalue).Take(&Child)
		db.Table("studentstatustable as childdiagnosis").Select("childdiagnosis.*").Where("childdiagnosis.StudentID = ?", Keyvalue).Take(&ChildDiagnosis)
		db.Table("studentstatustable as childdiagnosis").Select("childdiagnosis.*").Where("childdiagnosis.StudentID = ?", Keyvalue).Take(&ChildFamily)
		db.Table("studentstatustable as childdiagnosis").Select("childdiagnosis.Status").Where("childdiagnosis.StudentID = ?", Keyvalue).Take(&ChildAllData.Status)
		ChildAllData.Child = Child
		ChildAllData.ChildDiagnosis = ChildDiagnosis
		ChildAllData.ChildFamily = ChildFamily

		var FillStatus = make([]types.FillStatus, 0)
		db.Table("studentfillfinish").Select("SchoolYear as Year, Semester, FillTime, Finish").Where("StudentID = ?", Keyvalue).Find(&FillStatus)
		for j := 0; j < len(FillStatus); j++ {
			var Time = FillStatus[j].FillTime
			var Finish = FillStatus[j].Finish
			var FillYear = FillStatus[j].Year
			var FillSemester = FillStatus[j].Semester
			var EachFillData types.FillData
			if Finish == 1 {
				for FillTime := 1; FillTime <= Time; FillTime++ {
					db.Table("questionstoretable as filldata, questionnaireframework as questiondata").Select("filldata.SchoolYear as FillYear, filldata.Semester as FillSemester, filldata.FillTime, filldata.FillDate, questiondata.QuestionCode, questiondata.QuestionName").Where("filldata.FillTime = ?", FillTime).Where("filldata.SchoolYear = ?", FillYear).Where("filldata.Semester = ?", FillSemester).Where("filldata.QuestionCode = questiondata.QuestionCode").Take(&EachFillData)
					FillData = append(FillData, EachFillData)
				}
			} else if Finish == 0 {
				for FillTime := 1; FillTime < Time; FillTime++ {
					db.Table("questionstoretable as filldata, questionnaireframework as questiondata").Select("filldata.SchoolYear as FillYear, filldata.Semester as FillSemester, filldata.FillTime, filldata.FillDate, questiondata.QuestionCode, questiondata.QuestionName").Where("filldata.FillTime = ?", FillTime).Where("filldata.SchoolYear = ?", FillYear).Where("filldata.Semester = ?", FillSemester).Where("filldata.QuestionCode = questiondata.QuestionCode").Take(&EachFillData)
					FillData = append(FillData, EachFillData)
				}
			}
		}

		return ChildAllData, FillData
	} else if SearchType == "FillResult" {
		var QuestionFill types.QuestionFill
		var QuestionGrade types.QuestionGrade

		var Child types.Child
		var FillData types.FillData
		SearchKey := strings.Split(Keyvalue, "-")
		StudentID := SearchKey[0]
		FillYear := SearchKey[1]
		FillSemester := SearchKey[2]
		QuestionCode := SearchKey[3]
		FillTime := SearchKey[4]

		db.Table("studentschooltable as childbasic, userdatatable as user").Select("childbasic.* ,user.TeacherName as TeacherName").Where("childbasic.SchoolCode = user.SchoolCode").Where("childbasic.TeacherAccount = user.Username").Where("childbasic.StudentID = ?", StudentID).Take(&Child)
		db.Table("questionstoretable as filldata, questionnaireframework as questiondata").Select("filldata.SchoolYear as FillYear, filldata.Semester as FillSemester, filldata.FillTime, filldata.FillDate, questiondata.QuestionCode, questiondata.QuestionName").Where("filldata.StudentID = ?", StudentID).Where("filldata.SchoolYear = ?", FillYear).Where("filldata.Semester = ?", FillSemester).Where("filldata.FillTime = ?", FillTime).Where("filldata.QuestionCode = ? ", QuestionCode).Take(&FillData)
		QuestionFill.Child = Child
		QuestionFill.FillData = FillData

		var QuesitonContent = make([]types.QuesitonContent, 0)
		var QuestionBasicGrade = make([]types.QuestionBasicGrade, 0)
		var QuestionDetailGrade = make([]types.QuestionDetailGrade, 0)
		var QuestionResult = make([]types.QuestionResult, 0)
		db.Table("questionnairecontent, questionnaireframework").Select("DISTINCT questionnairecontent.QuestionCode, questionnairecontent.BigTopicNumber, questionnairecontent.BigTopicName, questionnaireframework.QuestionName").Where("questionnairecontent.QuestionCode = ?", QuestionCode).Where("questionnairecontent.QuestionCode = questionnaireframework.QuestionCode").Find(&QuesitonContent)
		for i := 0; i < len(QuesitonContent); i++ {
			var SmTopicData = make([]types.SmTopicData, 0)
			db.Table("questionnairecontent").Where("QuestionCode = ?", QuesitonContent[i].QuestionCode).Where("BigTopicNumber = ?", QuesitonContent[i].QuestionCode).Find(&SmTopicData)
			QuesitonContent[i].SmTopicData = SmTopicData
		}
		db.Table("questionbasicgrade").Where("QuestionCode = ?", QuestionCode).Where("StudentID = ?", StudentID).Where("SchoolYear = ?", FillYear).Where("Semester = ?", FillSemester).Where("FillTime = ?", FillTime).Find(&QuestionBasicGrade)
		db.Table("questiondetailgrade").Where("QuestionCode = ?", QuestionCode).Where("StudentID = ?", StudentID).Where("SchoolYear = ?", FillYear).Where("Semester = ?", FillSemester).Where("FillTime = ?", FillTime).Find(&QuestionDetailGrade)
		db.Table("questionstoretable").Select("BigTopicNumber, Value as ResultValue").Where("questionstoretable.QuestionCode = ?", QuestionCode).Where("questionstoretable.StudentID = ?", StudentID).Where("questionstoretable.SchoolYear = ?", FillYear).Where("questionstoretable.Semester = ?", FillSemester).Find(&QuestionResult)
		QuestionGrade.QuesitonContent = QuesitonContent
		QuestionGrade.QuestionBasicGrade = QuestionBasicGrade
		QuestionGrade.QuestionDetailGrade = QuestionDetailGrade
		QuestionGrade.QuestionResult = QuestionResult
		return QuestionFill, QuestionGrade

	} else {
		return false, nil
	}
}
