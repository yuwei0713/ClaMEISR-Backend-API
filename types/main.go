package types

type SearchUsers struct {
	Username      string `json:"Account"`
	TeacherName   string `json:"TeacherName"`
	SchoolCode    int    `json:"SchoolCode"`
	SchoolName    string `json:"SchoolName"`
	Separate      string `json:"Sperate"`
	Kindergarten  string `json:"Kindergarten"`
	Counseling    string `json:"Counseling"`
	RoutinesBased string `json:"RoutinesBased"`
}
type FrontendUsers struct {
	Account    string `json:"Account" gorm:"column:Username"`
	Username   string `json:"Username" gorm:"column:TeacherName"`
	SchoolCode string `json:"SchoolCode"`
	SchoolName string `json:"SchoolName"`
	IfFill     int    `json:"IfFill"`
	CreateAt   string `json:"CreateAt" gorm:"column:created_at"`
}
type BackendUsers struct {
	Account    string `json:"Account" gorm:"account"`
	Username   string `json:"username" gorm:"username"`
	SchoolCode string `json:"SchoolCode"`
	Password   string `json:"Password"`
	Permission int    `json:"Permission"`
}

type Schools struct {
	SchoolName string    `json:"SchoolName" gorm:"column:SchoolName"`
	SchoolCode int       `json:"SchoolCode" gprm:"column:SchoolCode"`
	Classes    []Classes `json:"Class" gorm:"-"`
}

type Classes struct {
	ClassName string `json:"ClassName" gorm:"column:ClassName"`
	ClassCode int    `json:"ClassCode" gorm:"column:ClassCode"`
}

type Child struct {
	Year        int    `json:"Year"`
	Semester    string `json:"Semester"`
	SchoolName  string `json:"SchoolName"`
	SchoolCode  int    `json:"SchoolCode" gprm:"column:SchoolCode"`
	ClassName   string `json:"ClassName"`
	ClassCode   int    `json:"ClassCode" gorm:"column:ClassCode"`
	StudentCode int    `json:"StudentCode"`
	StudentName string `json:"ChildName"`
	Gender      string `json:"Gender"`
	BirthDay    string `json:"BirthDay"`
	Age         int    `json:"Age"`
	TeacherName string `json:"TeacherName"`
	StudentID   string `json:"StudentID"`
}
type ChildDetail struct {
	Status         string `json:"Status"`
	Child          `json:"ChildBasic"`
	ChildDiagnosis `json:"Diagnosis,omitempty"`
	ChildFamily    `json:"Family"`
}
type ChildDiagnosis struct {
	Identites      string `json:"Identites,omitempty"`
	Proofs         string `json:"Proofs,omitempty"`
	Diagnosis      string `json:"Diagnosis,omitempty"`
	OtherDiagnosis string `json:"OtherDiagnosis,omitempty"`
	Note           string `json:"Note,omitempty"`
	Degree         string `json:"Degree,omitempty"`
	Placement      string `json:"Placement,omitempty"`
	Manual         string `json:"Manual,omitempty"`
}
type ChildFamily struct {
	Resident      string `json:"Resident"`
	Fstattend     string `gorm:"column:Fst-attend" json:"Fstattend"`
	Secattend     string `gorm:"column:Sec-attend" json:"Secattend,omitempty"`
	OtherResident string `json:"OtherResident,omitempty"`
}
type QuestionFill struct {
	Child    `json:"StudentData"`
	FillData `json:"FillStatus"`
}
type FillData struct {
	QuestionName string `json:"QuestionName"`
	FillTime     int    `json:"FillTime"`
	FillDate     string `json:"FillDate"`
	FillYear     int    `json:"FillYear"`
	FillSemester string `json:"FillSemester"`
	QuestionCode int    `json:"QuestionCode"`
}
type FillStatus struct {
	FillTime int
	Finish   int
	Year     int
	Semester string
}
type QuestionGrade struct {
	QuesitonContent     []QuesitonContent     `json:"QuesitonContent" gorm:"-"`
	QuestionBasicGrade  []QuestionBasicGrade  `json:"QuestionBasicGrade" gorm:"-"`
	QuestionDetailGrade []QuestionDetailGrade `json:"QuestionDetailGrade" gorm:"-"`
	QuestionResult      []QuestionResult      `json:"QuestionResult" gorm:"-"`
}
type QuestionManage struct {
	QuestionCode  int    `json:"QuestionCode"`
	QuestionName  string `json:"QuestionName"`
	TopicQuantity int    `gorm:"column:QuestionQuantity" json:"TopicQuantity"`
	SuitableAge   string `json:"SuitableAge"`
}
type QuesitonContent struct {
	QuestionCode   int           `json:"QuestionCode"`
	QuestionName   string        `json:"QuestionName"`
	BigTopicNumber int           `json:"BigTopicNumber"`
	BigTopicName   string        `json:"BigTopicName"`
	SmTopicData    []SmTopicData `json:"SmTopicData" gorm:"-"`
}
type SmTopicData struct {
	SmTopicNumber         int    `json:"SmTopicNumber"`
	SmTopicContent        string `json:"SmTopicContent"`
	QuesitonDetailContent `json:"QuesitonDetailContent,omitempty"`
}
type QuesitonDetailContent struct {
	OptionType       string `json:"OptionType,omitempty"`
	SuitableAge      int    `json:"SuitableAge,omitempty"`
	AdditionQuantity int    `json:"AdditionQuantity,omitempty"`
	AdditionTitle    string `json:"AdditionTitle,omitempty"`
	AddtitionContent string `json:"AddtitionContent,omitempty"`
	OptionQuantity   int    `json:"OptionQuantity,omitempty"`
	OptionValue      string `json:"OptionValue,omitempty"`
	OptionContent    string `json:"OptionContent,omitempty"`
}
type QuestionBasicGrade struct {
	BigTopicNumber       int     `json:"BigTopicNumber"`
	ThreePoint           int     `json:"ThreePoint"`
	FillByAge            int     `json:"FillByAge"`
	AgeProficientPercent float32 `json:"AgeProficientPercent"`
	FillByAll            int     `json:"FillByAll"`
	AllProficientPercent float32 `json:"AllProficientPercent"`
}
type QuestionDetailGrade struct {
	BigTopicNumber       int     `json:"BigTopicNumber"`
	Category             string  `json:"Category"`
	DetailName           string  `json:"DetailName"`
	ThreePoint           int     `json:"ThreePoint"`
	FillByAge            int     `json:"FillByAge"`
	AgeProficientPercent float32 `json:"AgeProficientPercent"`
	FillByAll            int     `json:"FillByAll"`
	AllProficientPercent float32 `json:"AllProficientPercent"`
}
type QuestionResult struct {
	BigTopicNumber int    `json:"BigTopicNumber"`
	ResultValue    string `json:"ResultValue"`
	// FillTeacher    string `json:"FillTeacher"`
}
