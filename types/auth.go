package types

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

//for 單一新增
type FrontendUsers_Register struct {
	Account    string `json:"Account"`
	SchoolCode string `json:"SchoolCode"`
	Password   string `json:"Password"`
	Quantity   int    `json:"Quantity"`
}
