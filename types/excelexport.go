package types

type ExcelExportData struct { //final struct
	ChildDetail   ChildDetail
	FillData      FillData
	QuestionGrade QuestionGrade
}

type Excel_FillStore struct {
	StudentID    string `json:"StudentID"`
	QuestionCode int    `json:"QuestionCode"`
	Year         int    `json:"SchoolYear" gorm:"column:SchoolYear"`
	Semester     string `json:"Semester"`
	FillTime     int    `json:"FillTime" gorm:"column:fill_time_adjusted"`
}

type Excel_SchoolData struct {
	SchoolName string `json:"SchoolName" gorm:"column:SchoolName"`
	SchoolCode int    `json:"SchoolCode" gprm:"column:SchoolCode"`
	ClassName  string `json:"ClassName" gorm:"column:ClassName"`
	ClassCode  int    `json:"ClassCode" gorm:"column:ClassCode"`
}
