package models

import (
	Routines "ginapi/routine"
	"ginapi/types"
	"strconv"
)

func QuestionManage() []types.QuestionManage {
	var db = Routines.MeisrDB
	var QuestionManage = make([]types.QuestionManage, 0)
	db.Table("questionnaireframework").Find(&QuestionManage)
	var MinAge = 0
	var MaxAge = 0
	for i := 0; i < len(QuestionManage); i++ {
		db.Table("questionnairecontent").Select("MIN(SuitableAge)").Where("QuestionCode = ?", QuestionManage[i].QuestionCode).Take(&MinAge)
		db.Table("questionnairecontent").Select("MAX(SuitableAge)").Where("QuestionCode = ?", QuestionManage[i].QuestionCode).Take(&MaxAge)
		var SuitableAge = strconv.Itoa(MinAge) + "~" + strconv.Itoa(MaxAge)
		QuestionManage[i].SuitableAge = SuitableAge
	}
	return QuestionManage
}
func DetailQuestionManage(QuestionCode int) []types.QuesitonContent {
	var db = Routines.MeisrDB
	var QuesitonContent = make([]types.QuesitonContent, 0)
	db.Table("questionnairecontent, questionnaireframework").Select("DISTINCT questionnairecontent.QuestionCode, questionnairecontent.BigTopicNumber, questionnairecontent.BigTopicName, questionnaireframework.QuestionName").Where("questionnairecontent.QuestionCode = ?", QuestionCode).Where("questionnairecontent.QuestionCode = questionnaireframework.QuestionCode").Find(&QuesitonContent)
	for i := 0; i < len(QuesitonContent); i++ {
		var SmTopicData = make([]types.SmTopicData, 0)
		db.Table("questionnairecontent").Where("QuestionCode = ?", QuesitonContent[i].QuestionCode).Where("BigTopicNumber = ?", QuesitonContent[i].BigTopicNumber).Find(&SmTopicData)
		QuesitonContent[i].SmTopicData = SmTopicData
	}
	return QuesitonContent
}
