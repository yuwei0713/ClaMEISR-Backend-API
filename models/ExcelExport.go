package models

import (
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func DefaultExport(Year string, Semester string) *excelize.File {
	//create Excel file, in excelize, create file the first sheet will create too
	ExcelFile := excelize.NewFile()
	//after finish this function, or error. Close Excel
	defer func() {
		if err := ExcelFile.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	//set main sheet name
	MainSheet := "總表" // set cell will use this variable
	ExcelFile.SetSheetName("Sheet1", MainSheet)

	var db = Routines.MeisrDB
	//get school count
	var TotalSchool = make([]types.Schools, 0)
	db.Table("schooltable").Select("DISTINCT SchoolName ,SchoolCode").Where("SchoolCode != 99").Find(&TotalSchool)

	var EachSchool types.Excel_SchoolData
	//get all school code without '其他'
	schoolrow, _ := db.Table("schooltable").Select("DISTINCT SchoolName ,SchoolCode").Where("SchoolCode != 99").Rows()
	YearAndSemesterCell := 1 //學年期行數Cell
	YearAndSemester := fmt.Sprintf("%s-%s", Year, Semester)
	ExcelFile.SetCellValue(MainSheet, fmt.Sprintf("A%d", YearAndSemesterCell), YearAndSemester) //A1

	AreaCell := 2 //first area start from 2
	MainArea := map[string]string{
		"B": "學校",
		"C": "班級",
		"D": "教師",
		"E": "學生",
		"F": "問卷",
		"G": "次數",
		"H": "日期",
	}
	for CellName, CellValue := range MainArea {
		ExcelFile.SetCellValue(MainSheet, fmt.Sprintf("%s%d", CellName, AreaCell), CellValue)
	}
	AreaCell++ // +1 start content
	for schoolrow.Next() {
		var Excel_FillStore types.Excel_FillStore
		db.ScanRows(schoolrow, &EachSchool) // Get School and Class
		questionstorerow, _ := db.Table("studentfillfinish as filldata").
			Joins("JOIN studentschooltable as student ON student.StudentID = filldata.StudentID").
			Select(`filldata.*,
			CASE 
            	WHEN filldata.Finish = 0 AND (filldata.FillTime - 1) > 0 
				THEN filldata.FillTime - 1
            ELSE filldata.FillTime
			END AS fill_time_adjusted`).
			Where("filldata.SchoolYear = ?", Year).Where("filldata.Semester = ?", Semester).
			Where("student.SchoolCode = ?", EachSchool.SchoolCode).Where("student.ClassCode = ?", EachSchool.ClassCode).Rows()
		for questionstorerow.Next() {
			var ExcelExportData types.ExcelExportData
			db.ScanRows(questionstorerow, &Excel_FillStore) // Get Questionstoretable data//
			for CurrentFillTime := 1; CurrentFillTime <= Excel_FillStore.FillTime; CurrentFillTime++ {
				SubSheetCell := 2
				//StudentID => Student Data (Basic , Status, Diagnosis)
				ChildData, _, _ := DetailSearch("Children", "", Excel_FillStore.StudentID)
				if childDetail, ok := ChildData.(types.ChildDetail); ok {
					ExcelExportData.ChildDetail = childDetail
				} else {
					// 如果類型不匹配，做相應的處理或報錯
					log.Println("ChildData is not of type ChildDetailType")
				}
				//Question Basic, Detail Data
				//keyvalue = StudentID - FillYear - FillSemester - QuestionCode - FillTime
				QuestionValue := fmt.Sprintf("%s-%d-%s-%d-%d", Excel_FillStore.StudentID, Excel_FillStore.Year, Excel_FillStore.Semester, Excel_FillStore.QuestionCode, CurrentFillTime)
				QuestionFill, QuestionGrade, _ := DetailSearch("FillResult", "", QuestionValue)
				if fillData, ok := QuestionFill.(types.QuestionFill); ok {
					ExcelExportData.FillData = fillData.FillData
				} else {
					log.Println("QuestionFill is not of type FillDataType")
				}
				if questionGrade, ok := QuestionGrade.(types.QuestionGrade); ok {
					ExcelExportData.QuestionGrade = questionGrade
				} else {
					log.Println("QuestionGrade is not of type QuestionGradeType")
				}
				//Get Data Finish
				/*
					A + YearAndSemesterCell 填寫學年期
					B ~ H + TitleCell 標題(學校，班級，教師，學生，問卷，次數，日期)
					(J)K ~ N + TotalCountCell 總數(學校名稱，班級數，教師數，學生數)，最後一行放入總和(學校數，班級數，教師數，學生數)
				*/
				ContentValue := [...]string{EachSchool.SchoolName, EachSchool.ClassName, ExcelExportData.ChildDetail.Child.TeacherName, ExcelExportData.ChildDetail.Child.StudentName, ExcelExportData.FillData.QuestionName, strconv.Itoa(ExcelExportData.FillData.FillTime), ExcelExportData.FillData.FillDate}
				i := 0
				for CellName := range MainArea {
					ExcelFile.SetCellValue(MainSheet, fmt.Sprintf("%s%d", CellName, AreaCell), ContentValue[i])
					i++
				}
				AreaCell++
				//Main Sheet Done, Create Questionnaire Grade Sheet
				QuestionnaireSheet := fmt.Sprintf("%s-%s-次數%d", ExcelExportData.ChildDetail.Child.StudentName, ExcelExportData.FillData.QuestionName, ExcelExportData.FillData.FillTime)
				ExcelFile.NewSheet(QuestionnaireSheet)
				var SubSheet_ChartCellStart int
				//Back to Main Sheet
				ExcelFile.SetCellValue(QuestionnaireSheet, "A1", "回總表")
				ExcelFile.SetCellHyperLink(QuestionnaireSheet, "A1", fmt.Sprintf("%s!A%d", "MainSheet", (AreaCell-1)), "Location")
				SubSheet_StudentArea := map[string]string{
					"A": "學校",
					"B": "班級",
					"C": "座號",
					"D": "姓名",
					"E": "性別",
					"F": "兒童狀態",
					"G": "診斷",
					"H": "生日",
					"I": "第幾次填寫",
					"J": "填寫日期",
				}
				for CellName, CellValue := range SubSheet_StudentArea {
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
				}
				SubSheetCell++
				var SubSheet_StudentContent []string
				SubSheet_StudentContent = append(SubSheet_StudentContent, EachSchool.SchoolName)
				SubSheet_StudentContent = append(SubSheet_StudentContent, EachSchool.ClassName)
				SubSheet_StudentContent = append(SubSheet_StudentContent, strconv.Itoa(ExcelExportData.ChildDetail.Child.StudentCode))
				SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.Child.StudentName)
				SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.Child.Gender)
				switch ExcelExportData.ChildDetail.Status {
				case "confirm":
					SubSheet_StudentContent = append(SubSheet_StudentContent, "特殊生")
					if ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis == "other" {
						SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.ChildDiagnosis.OtherDiagnosis)
					} else {
						SubSheet_StudentContent = append(SubSheet_StudentContent, fmt.Sprintf("%s:%s", ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis, ExcelExportData.ChildDetail.ChildDiagnosis.Degree))
					}
				case "suspected":
					SubSheet_StudentContent = append(SubSheet_StudentContent, "疑似生")
					if ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis == "other" {
						SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.ChildDiagnosis.OtherDiagnosis)
					} else {
						SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis)
					}
				case "none":
					SubSheet_StudentContent = append(SubSheet_StudentContent, "一般生")
					SubSheet_StudentContent = append(SubSheet_StudentContent, "無")
				}
				SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.Child.BirthDay)
				SubSheet_StudentContent = append(SubSheet_StudentContent, strconv.Itoa(ExcelExportData.FillData.FillTime))
				SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.FillData.FillDate)
				Contentlenth := 0
				for CellName := range SubSheet_StudentArea {
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_StudentContent[Contentlenth])
					Contentlenth++
				}
				SubSheetCell += 2
				//Main Area Basic Grade
				SubSheet_ChartCellStart = SubSheetCell
				SubSheet_MainArea := map[string]string{
					"A": "ClaMEISR 作息類別 (各類作息的題數)",
					"B": "作息被為評 3 分的題數",
					"C": "作息符合年齡的所有題數",
					"D": "符合年齡的精熟度",
					"E": "作息全部的題數",
					"F": "作息整體的精熟度",
				}
				for CellName, CellValue := range SubSheet_MainArea {
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
				}
				SubSheetCell++
				for TopicLength := 1; TopicLength <= len(ExcelExportData.QuestionGrade.QuestionBasicGrade); TopicLength++ {
					var SubSheet_BasicGrade []string
					if TopicLength == len(ExcelExportData.QuestionGrade.QuestionBasicGrade) { //use topiclength 0
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuesitonContent[0].BigTopicName)
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.Itoa(ExcelExportData.QuestionGrade.QuestionBasicGrade[0].ThreePoint))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.Itoa(ExcelExportData.QuestionGrade.QuestionBasicGrade[0].FillByAge))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.FormatFloat(float64(ExcelExportData.QuestionGrade.QuestionBasicGrade[0].AgeProficientPercent), 'f', -1, 32))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.Itoa(ExcelExportData.QuestionGrade.QuestionBasicGrade[0].FillByAll))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.FormatFloat(float64(ExcelExportData.QuestionGrade.QuestionBasicGrade[0].AllProficientPercent), 'f', -1, 32))
					} else {
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuesitonContent[TopicLength].BigTopicName)
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.Itoa(ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength].ThreePoint))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.Itoa(ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength].FillByAge))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.FormatFloat(float64(ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength].AgeProficientPercent), 'f', -1, 32))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.Itoa(ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength].FillByAll))
						SubSheet_BasicGrade = append(SubSheet_BasicGrade, strconv.FormatFloat(float64(ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength].AllProficientPercent), 'f', -1, 32))
					}
					Contentlenth := 0
					for CellName := range SubSheet_StudentArea {
						ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_BasicGrade[Contentlenth])
						Contentlenth++
					}
					SubSheetCell++
				}
				SubSheetCell++
				// SubSheet_MainChart
				ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", SubSheet_ChartCellStart), &excelize.Chart{
					Type: excelize.Line,
					Series: []excelize.ChartSeries{
						{
							Name:       fmt.Sprintf("%s!$D$%d", QuestionnaireSheet, SubSheet_ChartCellStart),
							Categories: fmt.Sprintf("%s!$A$%d:$A$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
							Values:     fmt.Sprintf("%s!$D$%d:$D$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
							Line: excelize.ChartLine{
								Smooth: true,
							},
						},
						{
							Name:       fmt.Sprintf("%s!$F$%d", QuestionnaireSheet, SubSheet_ChartCellStart),
							Categories: fmt.Sprintf("%s!$A$%d:$A$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
							Values:     fmt.Sprintf("%s!$F$%d:$F$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
							Line: excelize.ChartLine{
								Smooth: true,
							},
						},
					},
					Format: excelize.GraphicOptions{
						OffsetX: 15,
						OffsetY: 100,
					},
					Legend: excelize.ChartLegend{
						Position: "top",
					},
					Title: []excelize.RichTextRun{
						{
							Text: "問卷分數",
						},
					},
					PlotArea: excelize.ChartPlotArea{
						ShowCatName:     false,
						ShowLeaderLines: false,
						ShowPercent:     true,
						ShowSerName:     true,
						ShowVal:         true,
					},
					ShowBlanksAs: "zero",
				})

				// Classify Detail
				Func_E := make(map[int]types.QuestionDetailGrade)
				Func_I := make(map[int]types.QuestionDetailGrade)
				Func_SR := make(map[int]types.QuestionDetailGrade)
				Dev_A := make(map[int]types.QuestionDetailGrade)
				Dev_CG := make(map[int]types.QuestionDetailGrade)
				Dev_CM := make(map[int]types.QuestionDetailGrade)
				Dev_M := make(map[int]types.QuestionDetailGrade)
				Dev_S := make(map[int]types.QuestionDetailGrade)
				Out_One := make(map[int]types.QuestionDetailGrade)
				Out_Two := make(map[int]types.QuestionDetailGrade)
				Out_Three := make(map[int]types.QuestionDetailGrade)
				for DetailTopicLength := 0; DetailTopicLength < len(ExcelExportData.QuestionGrade.QuestionDetailGrade); DetailTopicLength++ {
					switch ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].Category {
					case "Func":
						switch ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].DetailName {
						case "E":
							Func_E[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "I":
							Func_I[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "SR":
							Func_SR[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						}
					case "Dev":
						switch ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].DetailName {
						case "A":
							Dev_A[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "CG":
							Dev_CG[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "CM":
							Dev_CM[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "M":
							Dev_M[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "S":
							Dev_S[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						}
					case "Out":
						switch ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].DetailName {
						case "1":
							Out_One[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "2":
							Out_Two[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						case "3":
							Out_Three[ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength].BigTopicNumber] = ExcelExportData.QuestionGrade.QuestionDetailGrade[DetailTopicLength]
						}
					}
				}
				// SubSheet_FuncArea
				SubSheet_ChartCellStart = SubSheetCell
				SubSheet_FuncArea := map[string]string{
					"A": "ClaMEISR 作息類別 (各類作息的題數)",
					"B": "成效領域名稱",
					"C": "作息被為評 3 分的題數",
					"D": "作息符合年齡的所有題數",
					"E": "符合年齡的精熟度",
					"F": "作息全部的題數",
					"H": "作息整體的精熟度",
				}
				for CellName, CellValue := range SubSheet_FuncArea {
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
				}
				SubSheetCell++
				for FuncTopicLength := 1; FuncTopicLength <= len(Func_E); FuncTopicLength++ {
					if FuncTopicLength == len(Func_E) {
						FuncTopicLength = 0
					}
					TopicStartCell := SubSheetCell
					var SubSheet_Func_E_Grade []string
					var Contentlenth = 0
					SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].DetailName)
					SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, strconv.Itoa(Func_E[FuncTopicLength].ThreePoint))
					SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, strconv.Itoa(Func_E[FuncTopicLength].FillByAge))
					SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, strconv.FormatFloat(float64(Func_E[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, strconv.Itoa(Func_E[FuncTopicLength].FillByAll))
					SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, strconv.FormatFloat(float64(Func_E[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Func_E_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Func_I_Grade []string
					SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].DetailName)
					SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, strconv.Itoa(Func_I[FuncTopicLength].ThreePoint))
					SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, strconv.Itoa(Func_I[FuncTopicLength].FillByAge))
					SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, strconv.FormatFloat(float64(Func_I[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, strconv.Itoa(Func_I[FuncTopicLength].FillByAll))
					SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, strconv.FormatFloat(float64(Func_I[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Func_I_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Func_SR_Grade []string
					SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].DetailName)
					SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, strconv.Itoa(Func_SR[FuncTopicLength].ThreePoint))
					SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, strconv.Itoa(Func_SR[FuncTopicLength].FillByAge))
					SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, strconv.FormatFloat(float64(Func_SR[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, strconv.Itoa(Func_SR[FuncTopicLength].FillByAll))
					SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, strconv.FormatFloat(float64(Func_SR[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Func_SR_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					TopicEndCell := SubSheetCell
					//Combine A Cell
					ExcelFile.MergeCell(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), fmt.Sprintf("A%d", TopicEndCell))
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), ExcelExportData.QuestionGrade.QuesitonContent[FuncTopicLength].QuestionName)
					SubSheetCell++
				}
				SubSheetCell++
				// SubSheet_FuncChart
				ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", SubSheet_ChartCellStart), &excelize.Chart{
					Type: excelize.Bar,
					Series: []excelize.ChartSeries{
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 6)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 6), (SubSheetCell - 6)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 5)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 5), (SubSheetCell - 5)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 4)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 4)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 3)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 3), (SubSheetCell - 3)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 2)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 2), (SubSheetCell - 2)),
						},
					},
					Format: excelize.GraphicOptions{
						OffsetX: 15,
						OffsetY: 20,
					},
					Legend: excelize.ChartLegend{
						Position: "left",
					},
					Title: []excelize.RichTextRun{
						{
							Text: "Func Chart",
						},
					},
					PlotArea: excelize.ChartPlotArea{
						ShowCatName:     false,
						ShowLeaderLines: false,
						ShowPercent:     true,
						ShowSerName:     true,
						ShowVal:         true,
					},
					ShowBlanksAs: "zero",
				})
				// SubSheet_DevArea
				SubSheet_ChartCellStart = SubSheetCell
				for CellName, CellValue := range SubSheet_FuncArea {
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
				}
				SubSheetCell++
				for FuncTopicLength := 1; FuncTopicLength <= len(Dev_A); FuncTopicLength++ {
					if FuncTopicLength == len(Dev_A) {
						FuncTopicLength = 0
					}
					TopicStartCell := SubSheetCell
					var SubSheet_Dev_A_Grade []string
					var Contentlenth = 0
					SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[FuncTopicLength].DetailName)
					SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, strconv.Itoa(Dev_A[FuncTopicLength].ThreePoint))
					SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, strconv.Itoa(Dev_A[FuncTopicLength].FillByAge))
					SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, strconv.FormatFloat(float64(Dev_A[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, strconv.Itoa(Dev_A[FuncTopicLength].FillByAll))
					SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, strconv.FormatFloat(float64(Dev_A[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_A_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Dev_CG_Grade []string
					SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[FuncTopicLength].DetailName)
					SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, strconv.Itoa(Dev_CG[FuncTopicLength].ThreePoint))
					SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, strconv.Itoa(Dev_CG[FuncTopicLength].FillByAge))
					SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, strconv.FormatFloat(float64(Dev_CG[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, strconv.Itoa(Dev_CG[FuncTopicLength].FillByAll))
					SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, strconv.FormatFloat(float64(Dev_CG[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_CG_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Dev_CM_Grade []string
					SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[FuncTopicLength].DetailName)
					SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, strconv.Itoa(Dev_CM[FuncTopicLength].ThreePoint))
					SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, strconv.Itoa(Dev_CM[FuncTopicLength].FillByAge))
					SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, strconv.FormatFloat(float64(Dev_CM[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, strconv.Itoa(Dev_CM[FuncTopicLength].FillByAll))
					SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, strconv.FormatFloat(float64(Dev_CM[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_CM_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Dev_M_Grade []string
					SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[FuncTopicLength].DetailName)
					SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, strconv.Itoa(Dev_M[FuncTopicLength].ThreePoint))
					SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, strconv.Itoa(Dev_M[FuncTopicLength].FillByAge))
					SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, strconv.FormatFloat(float64(Dev_M[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, strconv.Itoa(Dev_M[FuncTopicLength].FillByAll))
					SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, strconv.FormatFloat(float64(Dev_M[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_M_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Dev_S_Grade []string
					SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[FuncTopicLength].DetailName)
					SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, strconv.Itoa(Dev_S[FuncTopicLength].ThreePoint))
					SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, strconv.Itoa(Dev_S[FuncTopicLength].FillByAge))
					SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, strconv.FormatFloat(float64(Dev_S[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, strconv.Itoa(Dev_S[FuncTopicLength].FillByAll))
					SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, strconv.FormatFloat(float64(Dev_S[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_S_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					TopicEndCell := SubSheetCell
					//Combine A Cell
					ExcelFile.MergeCell(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), fmt.Sprintf("A%d", TopicEndCell))
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), ExcelExportData.QuestionGrade.QuesitonContent[FuncTopicLength].QuestionName)
					SubSheetCell++
				}
				SubSheetCell++
				// SubSheet_DevChart
				ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", SubSheet_ChartCellStart), &excelize.Chart{
					Type: excelize.Bar,
					Series: []excelize.ChartSeries{
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 4)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 4)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 3)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 3), (SubSheetCell - 3)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 2)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 2), (SubSheetCell - 2)),
						},
					},
					Format: excelize.GraphicOptions{
						OffsetX: 15,
						OffsetY: 15,
					},
					Legend: excelize.ChartLegend{
						Position: "left",
					},
					Title: []excelize.RichTextRun{
						{
							Text: "Dev Chart",
						},
					},
					PlotArea: excelize.ChartPlotArea{
						ShowCatName:     false,
						ShowLeaderLines: false,
						ShowPercent:     true,
						ShowSerName:     true,
						ShowVal:         true,
					},
					ShowBlanksAs: "zero",
				})
				// SubSheet_OutArea
				SubSheet_ChartCellStart = SubSheetCell
				for CellName, CellValue := range SubSheet_FuncArea {
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
				}
				SubSheetCell++
				for FuncTopicLength := 1; FuncTopicLength <= len(Out_One); FuncTopicLength++ {
					if FuncTopicLength == len(Out_One) {
						FuncTopicLength = 0
					}
					TopicStartCell := SubSheetCell
					var SubSheet_Out_One_Grade []string
					var Contentlenth = 0
					SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[FuncTopicLength].DetailName)
					SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, strconv.Itoa(Out_One[FuncTopicLength].ThreePoint))
					SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, strconv.Itoa(Out_One[FuncTopicLength].FillByAge))
					SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, strconv.FormatFloat(float64(Out_One[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, strconv.Itoa(Out_One[FuncTopicLength].FillByAll))
					SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, strconv.FormatFloat(float64(Out_One[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Out_One_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Out_Two_Grade []string
					SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[FuncTopicLength].DetailName)
					SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, strconv.Itoa(Out_Two[FuncTopicLength].ThreePoint))
					SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, strconv.Itoa(Out_Two[FuncTopicLength].FillByAge))
					SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, strconv.FormatFloat(float64(Out_Two[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, strconv.Itoa(Out_Two[FuncTopicLength].FillByAll))
					SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, strconv.FormatFloat(float64(Out_Two[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Out_Two_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					var SubSheet_Out_Three_Grade []string
					SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[FuncTopicLength].DetailName)
					SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, strconv.Itoa(Out_Three[FuncTopicLength].ThreePoint))
					SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, strconv.Itoa(Out_Three[FuncTopicLength].FillByAge))
					SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, strconv.FormatFloat(float64(Out_Three[FuncTopicLength].AgeProficientPercent), 'f', -1, 32))
					SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, strconv.Itoa(Out_Three[FuncTopicLength].FillByAll))
					SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, strconv.FormatFloat(float64(Out_Three[FuncTopicLength].AllProficientPercent), 'f', -1, 32))
					Contentlenth = 0
					for CellName := range SubSheet_FuncArea {
						if CellName != "A" {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Out_Three_Grade[Contentlenth])
							Contentlenth++
						}
					}
					SubSheetCell++
					TopicEndCell := SubSheetCell
					//Combine A Cell
					ExcelFile.MergeCell(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), fmt.Sprintf("A%d", TopicEndCell))
					ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), ExcelExportData.QuestionGrade.QuesitonContent[FuncTopicLength].QuestionName)
					SubSheetCell++
				}
				// SubSheet_OutChart
				ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", SubSheet_ChartCellStart), &excelize.Chart{
					Type: excelize.Bar,
					Series: []excelize.ChartSeries{
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 4)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 4)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 3)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 3), (SubSheetCell - 3)),
						},
						{
							Name:       fmt.Sprintf("%s!$B$%d", QuestionnaireSheet, (SubSheetCell - 2)),
							Categories: fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart), (SubSheet_ChartCellStart)),
							Values:     fmt.Sprintf("%s!$E$%d,$G$%d", QuestionnaireSheet, (SubSheetCell - 2), (SubSheetCell - 2)),
						},
					},
					Format: excelize.GraphicOptions{
						OffsetX: 15,
						OffsetY: 20,
					},
					Legend: excelize.ChartLegend{
						Position: "left",
					},
					Title: []excelize.RichTextRun{
						{
							Text: "Out Chart",
						},
					},
					PlotArea: excelize.ChartPlotArea{
						ShowCatName:     false,
						ShowLeaderLines: false,
						ShowPercent:     true,
						ShowSerName:     true,
						ShowVal:         true,
					},
					ShowBlanksAs: "zero",
				})
			}

		}
	}

	return ExcelFile
}

func CustomizeExport() {

}
