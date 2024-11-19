package models

import (
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"
	"log"
	"strconv"
	"time"

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
	db.Table("schooltable").Select("DISTINCT SchoolName ,SchoolCode").Where("SchoolCode != 99").Find(&TotalSchool) //學校名稱以及學校代碼的對照表

	schoolrow, _ := db.Table("schooltable").Select("*").Where("SchoolCode != 99").Rows()        // 學校每個班級代碼以及相對應的學校代碼跟班級名稱
	YearAndSemesterCell := 1                                                                    //學年期行數Cell
	YearAndSemester := fmt.Sprintf("%s-%s", Year, Semester)                                     //整合學年學期
	ExcelFile.SetCellValue(MainSheet, fmt.Sprintf("A%d", YearAndSemesterCell), YearAndSemester) //A1,設定在第A欄第一列

	AreaCell := 2                  //從第二裂開始寫
	MainArea := map[string]string{ // 處理execl 欄位(從第B欄開始)以及順序
		"B": "學校",
		"C": "班級",
		"D": "座號",
		"E": "教師",
		"F": "學生",
		"G": "性別",
		"H": "兒童狀態",
		"I": "診斷",
		"J": "生日",
		"K": "第幾次填寫",
		"L": "填寫日期",
		"M": "問卷",
		"N": "次數",
		"O": "日期",
	}
	MainArea_orderedKeys := []string{"B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O"}
	for _, CellName := range MainArea_orderedKeys {
		CellValue := MainArea[CellName]                                                       //對應的中文標籤
		ExcelFile.SetCellValue(MainSheet, fmt.Sprintf("%s%d", CellName, AreaCell), CellValue) //生成儲存格的地址(第X欄第X列)
	}
	var EachSchool types.Excel_SchoolData
	//get all school code without '其他'
	AreaCell++ // +1 start content
	for schoolrow.Next() {
		var Excel_FillStore types.Excel_FillStore
		db.ScanRows(schoolrow, &EachSchool) //**** Get School and Class
		query := ` 
			SELECT filldata.*,
			    CASE 
			        WHEN filldata.Finish = 0 AND filldata.FillTime > 1
			        THEN filldata.FillTime - 1
			        ELSE filldata.FillTime
			    END AS fill_time_adjusted
			FROM studentfillfinish as filldata
			JOIN studentschooltable as student ON student.StudentID = filldata.StudentID
			WHERE filldata.SchoolYear = ?
			    AND filldata.Semester = ?
			    AND student.SchoolCode = ?
			    AND student.ClassCode = ?
				AND NOT (filldata.FillTime = 1 AND filldata.Finish = 0)`
		//把fillfinish 跟 schootable 做 join，把兩者資料整合(問卷填寫+學校班級學生資料)
		questionstorerow, err := db.Raw(query, Year, Semester, EachSchool.SchoolCode, EachSchool.ClassCode).Rows()
		if err != nil {
			// 處理錯誤
			log.Println("錯誤：", err)
		}
		for questionstorerow.Next() {
			var ExcelExportData types.ExcelExportData
			db.ScanRows(questionstorerow, &Excel_FillStore) // Get Questionstoretable data//
			if Excel_FillStore.FillTime > 0 {               //確定該學生有
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

					layout := time.RFC3339 // 標準時間格式
					t, err := time.Parse(layout, ExcelExportData.FillData.FillDate)
					if err != nil {
						fmt.Println("Error parsing date:", err)
					}
					switch ExcelExportData.ChildDetail.Status {
					case "confirm":
						ExcelExportData.ChildDetail.Status = "特殊生"
					case "suspected":
						ExcelExportData.ChildDetail.Status = "疑似生"
					case "none":
						ExcelExportData.ChildDetail.Status = "一般生"
					}
					Date_Output := t.Format("2006-01-02") // 轉換為 YYYY-MM-DD 格式
					ContentValue := [...]string{
						EachSchool.SchoolName, // B: 學校
						EachSchool.ClassName,  // C: 班級
						strconv.Itoa(ExcelExportData.ChildDetail.Child.StudentCode), // D: 座號
						ExcelExportData.ChildDetail.Child.TeacherName,               // E: 教師
						ExcelExportData.ChildDetail.Child.StudentName,               // F: 學生
						ExcelExportData.ChildDetail.Child.Gender,                    // G: 性別
						ExcelExportData.ChildDetail.Status,                          // H: 兒童狀態
						ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis,        // I: 診斷
						ExcelExportData.ChildDetail.Child.BirthDay,                  // J: 生日
						strconv.Itoa(CurrentFillTime),                               // K: 第幾次填寫
						ExcelExportData.FillData.FillDate,                           // L: 填寫日期
						ExcelExportData.FillData.QuestionName,                       // M: 問卷
						strconv.Itoa(Excel_FillStore.FillTime),                      // N: 次數
						Date_Output,                                                 // O: 日期
					}
					i := 0
					for _, CellName := range MainArea_orderedKeys {
						ExcelFile.SetCellValue(MainSheet, fmt.Sprintf("%s%d", CellName, AreaCell), ContentValue[i])
						i++
					}
					AreaCell++
					//Main Sheet Done, Create Questionnaire Grade Sheet
					QuestionnaireSheet := fmt.Sprintf("%s-%s-次數%d", ExcelExportData.ChildDetail.Child.StudentName, ExcelExportData.FillData.QuestionName, ExcelExportData.FillData.FillTime)
					ExcelFile.NewSheet(QuestionnaireSheet)
					ExcelFile.SetCellHyperLink(MainSheet, fmt.Sprintf("G%d", AreaCell-1), fmt.Sprintf("'%s'!A1", QuestionnaireSheet), "Location")
					ExcelFile.SetCellHyperLink(MainSheet, fmt.Sprintf("F%d", AreaCell-1), fmt.Sprintf("'%s'!A1", QuestionnaireSheet), "Location")
					var SubSheet_ChartCellStart int
					//Back to Main Sheet
					ExcelFile.SetCellValue(QuestionnaireSheet, "A1", "回總表")
					ExcelFile.SetCellHyperLink(QuestionnaireSheet, "A1", fmt.Sprintf("%s!A%d", "總表", (AreaCell-1)), "Location")
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
					SubSheet_StudentArea_OrderedKeys := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
					for _, CellName := range SubSheet_StudentArea_OrderedKeys {
						CellValue := SubSheet_StudentArea[CellName]
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
					case "特殊生":
						SubSheet_StudentContent = append(SubSheet_StudentContent, "特殊生")
						if ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis == "other" {
							SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.ChildDiagnosis.OtherDiagnosis)
						} else {
							SubSheet_StudentContent = append(SubSheet_StudentContent, fmt.Sprintf("%s:%s", ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis, ExcelExportData.ChildDetail.ChildDiagnosis.Degree))
						}
					case "疑似生":
						SubSheet_StudentContent = append(SubSheet_StudentContent, "疑似生")
						if ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis == "other" {
							SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.ChildDiagnosis.OtherDiagnosis)
						} else {
							SubSheet_StudentContent = append(SubSheet_StudentContent, ExcelExportData.ChildDetail.ChildDiagnosis.Diagnosis)
						}
					case "一般生":
						SubSheet_StudentContent = append(SubSheet_StudentContent, "一般生")
						SubSheet_StudentContent = append(SubSheet_StudentContent, "無")
					}
					t, _ = time.Parse(layout, ExcelExportData.ChildDetail.Child.BirthDay)
					BirthDay_Output := t.Format("2006-01-02") // 轉換為 YYYY-MM-DD 格式
					SubSheet_StudentContent = append(SubSheet_StudentContent, BirthDay_Output)
					SubSheet_StudentContent = append(SubSheet_StudentContent, strconv.Itoa(ExcelExportData.FillData.FillTime))
					SubSheet_StudentContent = append(SubSheet_StudentContent, Date_Output)
					Contentlenth := 0
					for _, CellName := range SubSheet_StudentArea_OrderedKeys {
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

					SubSheet_MainArea_OrderedKeys := []string{"A", "B", "C", "D", "E", "F"} // 指定鍵的順序

					for _, CellName := range SubSheet_MainArea_OrderedKeys {
						CellValue := SubSheet_MainArea[CellName]
						ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
					}
					SubSheetCell++
					for TopicLength := 0; TopicLength < len(ExcelExportData.QuestionGrade.QuestionBasicGrade); TopicLength++ {
						var SubSheet_BasicGrade []any
						if TopicLength == (len(ExcelExportData.QuestionGrade.QuestionBasicGrade) - 1) { //use topiclength 0
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, "總和")
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[0].ThreePoint)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[0].FillByAge)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[0].AgeProficientPercent)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[0].FillByAll)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[0].AllProficientPercent)
						} else {
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuesitonContent[TopicLength].BigTopicName)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength+1].ThreePoint)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength+1].FillByAge)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength+1].AgeProficientPercent)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength+1].FillByAll)
							SubSheet_BasicGrade = append(SubSheet_BasicGrade, ExcelExportData.QuestionGrade.QuestionBasicGrade[TopicLength+1].AllProficientPercent)
						}
						Contentlenth := 0
						for _, CellName := range SubSheet_MainArea_OrderedKeys {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_BasicGrade[Contentlenth])
							Contentlenth++
						}
						SubSheetCell++
					}
					SubSheetCell++
					// SubSheet_MainChart
					if err := ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", SubSheet_ChartCellStart), &excelize.Chart{
						Type: excelize.Col,
						Series: []excelize.ChartSeries{
							{
								Name:       fmt.Sprintf("'%s'!$D$%d", QuestionnaireSheet, SubSheet_ChartCellStart),
								Categories: fmt.Sprintf("'%s'!$A$%d:$A$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$D$%d:$D$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
								Line: excelize.ChartLine{
									Smooth: true,
								},
							},
							{
								Name:       fmt.Sprintf("'%s'!$F$%d", QuestionnaireSheet, SubSheet_ChartCellStart),
								Categories: fmt.Sprintf("'%s'!$A$%d:$A$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$F$%d:$F$%d", QuestionnaireSheet, (SubSheet_ChartCellStart + 1), (SubSheetCell - 2)),
								Line: excelize.ChartLine{
									Smooth: true,
								},
							},
						},
						Format: excelize.GraphicOptions{
							ScaleX: 2.5,
							ScaleY: 1.5,
						},
						Legend: excelize.ChartLegend{
							Position: "top",
						},
						Title: []excelize.RichTextRun{
							{
								Text: "問卷分數",
							},
						},
						XAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   true,
								Italic: true,
								Color:  "#000000",
							},
						},
						YAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   false,
								Italic: false,
								Color:  "#777777",
							},
						},
						PlotArea: excelize.ChartPlotArea{
							ShowCatName:     false,
							ShowLeaderLines: false,
							ShowPercent:     true,
							ShowSerName:     false,
							ShowVal:         true,
						},
						ShowBlanksAs: "gap",
					}); err != nil {
						fmt.Println(err)
					}
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
					SubSheet_DetailArea := map[string]string{
						"A": "ClaMEISR 作息類別 (各類作息的題數)",
						"B": "成效領域名稱",
						"C": "作息被為評 3 分的題數",
						"D": "作息符合年齡的所有題數",
						"E": "符合年齡的精熟度",
						"F": "作息全部的題數",
						"G": "作息整體的精熟度",
					}

					SubSheet_DetailArea_OrderedKeys := []string{"A", "B", "C", "D", "E", "F", "G"} // 指定有序的鍵列表

					for _, CellName := range SubSheet_DetailArea_OrderedKeys {
						CellValue := SubSheet_DetailArea[CellName]
						ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
					}
					SubSheetCell++
					for FuncTopicLength := 1; FuncTopicLength <= len(Func_E); FuncTopicLength++ {
						if FuncTopicLength == len(Func_E) {
							FuncTopicLength = 0
						}
						TopicStartCell := SubSheetCell
						var SubSheet_Func_E_Grade []any
						var Contentlenth = 0
						SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].DetailName)
						SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].ThreePoint)
						SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].FillByAge)
						SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].AgeProficientPercent)
						SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].FillByAll)
						SubSheet_Func_E_Grade = append(SubSheet_Func_E_Grade, Func_E[FuncTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Func_E_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Func_I_Grade []any
						SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].DetailName)
						SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].ThreePoint)
						SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].FillByAge)
						SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].AgeProficientPercent)
						SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].FillByAll)
						SubSheet_Func_I_Grade = append(SubSheet_Func_I_Grade, Func_I[FuncTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Func_I_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Func_SR_Grade []any
						SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].DetailName)
						SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].ThreePoint)
						SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].FillByAge)
						SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].AgeProficientPercent)
						SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].FillByAll)
						SubSheet_Func_SR_Grade = append(SubSheet_Func_SR_Grade, Func_SR[FuncTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Func_SR_Grade[Contentlenth])
								Contentlenth++
							}
						}
						TopicEndCell := SubSheetCell
						//Combine A Cell
						ExcelFile.MergeCell(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), fmt.Sprintf("A%d", TopicEndCell))
						if FuncTopicLength != 0 {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), ExcelExportData.QuestionGrade.QuesitonContent[FuncTopicLength-1].BigTopicName)
						} else {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), "總和")
							SubSheetCell++
							break
						}
						SubSheetCell++
					}
					SubSheetCell++
					// SubSheet_FuncChart
					ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", (SubSheet_ChartCellStart+7)), &excelize.Chart{
						Type: excelize.Col,
						Series: []excelize.ChartSeries{
							{
								Name:       fmt.Sprintf("'%s'!$E$%d", QuestionnaireSheet, (SubSheet_ChartCellStart)),
								Categories: fmt.Sprintf("'%s'!$B$%d:$B$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$E$%d:$E$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
							},
							{
								Name:       fmt.Sprintf("'%s'!$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart)),
								Categories: fmt.Sprintf("'%s'!$B$%d:$B$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$G$%d:$G$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
							},
						},
						Format: excelize.GraphicOptions{
							ScaleX: 1.5,
							ScaleY: 1.5,
						},
						Legend: excelize.ChartLegend{
							Position: "top",
						},
						Title: []excelize.RichTextRun{
							{
								Text: "Func Chart",
							},
						},
						XAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   true,
								Italic: true,
								Color:  "#000000",
							},
						},
						YAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   false,
								Italic: false,
								Color:  "#777777",
							},
						},
						PlotArea: excelize.ChartPlotArea{
							ShowCatName:     false,
							ShowLeaderLines: false,
							ShowPercent:     true,
							ShowSerName:     false,
							ShowVal:         true,
						},
						ShowBlanksAs: "zero",
					})
					// SubSheet_DevArea
					SubSheet_ChartCellStart = SubSheetCell
					for _, CellName := range SubSheet_DetailArea_OrderedKeys {
						CellValue := SubSheet_DetailArea[CellName]
						ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
					}
					SubSheetCell++
					for DevTopicLength := 1; DevTopicLength <= len(Dev_A); DevTopicLength++ {
						if DevTopicLength == len(Dev_A) {
							DevTopicLength = 0
						}
						TopicStartCell := SubSheetCell
						var SubSheet_Dev_A_Grade []any
						var Contentlenth = 0
						SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[DevTopicLength].DetailName)
						SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[DevTopicLength].ThreePoint)
						SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[DevTopicLength].FillByAge)
						SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[DevTopicLength].AgeProficientPercent)
						SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[DevTopicLength].FillByAll)
						SubSheet_Dev_A_Grade = append(SubSheet_Dev_A_Grade, Dev_A[DevTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_A_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Dev_CG_Grade []any
						SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[DevTopicLength].DetailName)
						SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[DevTopicLength].ThreePoint)
						SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[DevTopicLength].FillByAge)
						SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[DevTopicLength].AgeProficientPercent)
						SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[DevTopicLength].FillByAll)
						SubSheet_Dev_CG_Grade = append(SubSheet_Dev_CG_Grade, Dev_CG[DevTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_CG_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Dev_CM_Grade []any
						SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[DevTopicLength].DetailName)
						SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[DevTopicLength].ThreePoint)
						SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[DevTopicLength].FillByAge)
						SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[DevTopicLength].AgeProficientPercent)
						SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[DevTopicLength].FillByAll)
						SubSheet_Dev_CM_Grade = append(SubSheet_Dev_CM_Grade, Dev_CM[DevTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_CM_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Dev_M_Grade []any
						SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[DevTopicLength].DetailName)
						SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[DevTopicLength].ThreePoint)
						SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[DevTopicLength].FillByAge)
						SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[DevTopicLength].AgeProficientPercent)
						SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[DevTopicLength].FillByAll)
						SubSheet_Dev_M_Grade = append(SubSheet_Dev_M_Grade, Dev_M[DevTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_M_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Dev_S_Grade []any
						SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[DevTopicLength].DetailName)
						SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[DevTopicLength].ThreePoint)
						SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[DevTopicLength].FillByAge)
						SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[DevTopicLength].AgeProficientPercent)
						SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[DevTopicLength].FillByAll)
						SubSheet_Dev_S_Grade = append(SubSheet_Dev_S_Grade, Dev_S[DevTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Dev_S_Grade[Contentlenth])
								Contentlenth++
							}
						}

						TopicEndCell := SubSheetCell
						//Combine A Cell
						ExcelFile.MergeCell(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), fmt.Sprintf("A%d", TopicEndCell))
						if DevTopicLength != 0 {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), ExcelExportData.QuestionGrade.QuesitonContent[DevTopicLength-1].BigTopicName)
						} else {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), "總和")
							SubSheetCell++
							break
						}
						SubSheetCell++
					}
					SubSheetCell++
					// SubSheet_DevChart
					ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", (SubSheet_ChartCellStart+7)), &excelize.Chart{
						Type: excelize.Col,
						Series: []excelize.ChartSeries{
							{
								Name:       fmt.Sprintf("'%s'!$E$%d", QuestionnaireSheet, (SubSheet_ChartCellStart)),
								Categories: fmt.Sprintf("'%s'!$B$%d:$B$%d", QuestionnaireSheet, (SubSheetCell - 6), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$E$%d:$E$%d", QuestionnaireSheet, (SubSheetCell - 6), (SubSheetCell - 2)),
							},
							{
								Name:       fmt.Sprintf("'%s'!$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart)),
								Categories: fmt.Sprintf("'%s'!$B$%d:$B$%d", QuestionnaireSheet, (SubSheetCell - 6), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$G$%d:$G$%d", QuestionnaireSheet, (SubSheetCell - 6), (SubSheetCell - 2)),
							},
						},
						Format: excelize.GraphicOptions{
							ScaleX: 2,
							ScaleY: 1.5,
						},
						Legend: excelize.ChartLegend{
							Position: "top",
						},
						Title: []excelize.RichTextRun{
							{
								Text: "Dev Chart",
							},
						},
						XAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   true,
								Italic: true,
								Color:  "#000000",
							},
						},
						YAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   false,
								Italic: false,
								Color:  "#777777",
							},
						},
						PlotArea: excelize.ChartPlotArea{
							ShowCatName:     false,
							ShowLeaderLines: false,
							ShowPercent:     true,
							ShowSerName:     false,
							ShowVal:         true,
						},
						ShowBlanksAs: "zero",
					})
					// SubSheet_OutArea
					SubSheet_ChartCellStart = SubSheetCell
					for _, CellName := range SubSheet_DetailArea_OrderedKeys {
						CellValue := SubSheet_DetailArea[CellName]
						ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), CellValue)
					}
					SubSheetCell++
					for OutTopicLength := 1; OutTopicLength <= len(Out_One); OutTopicLength++ {
						if OutTopicLength == len(Out_One) {
							OutTopicLength = 0
						}
						TopicStartCell := SubSheetCell
						var SubSheet_Out_One_Grade []any
						var Contentlenth = 0
						SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[OutTopicLength].DetailName)
						SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[OutTopicLength].ThreePoint)
						SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[OutTopicLength].FillByAge)
						SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[OutTopicLength].AgeProficientPercent)
						SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[OutTopicLength].FillByAll)
						SubSheet_Out_One_Grade = append(SubSheet_Out_One_Grade, Out_One[OutTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Out_One_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Out_Two_Grade []any
						SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[OutTopicLength].DetailName)
						SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[OutTopicLength].ThreePoint)
						SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[OutTopicLength].FillByAge)
						SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[OutTopicLength].AgeProficientPercent)
						SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[OutTopicLength].FillByAll)
						SubSheet_Out_Two_Grade = append(SubSheet_Out_Two_Grade, Out_Two[OutTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Out_Two_Grade[Contentlenth])
								Contentlenth++
							}
						}
						SubSheetCell++
						var SubSheet_Out_Three_Grade []any
						SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[OutTopicLength].DetailName)
						SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[OutTopicLength].ThreePoint)
						SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[OutTopicLength].FillByAge)
						SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[OutTopicLength].AgeProficientPercent)
						SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[OutTopicLength].FillByAll)
						SubSheet_Out_Three_Grade = append(SubSheet_Out_Three_Grade, Out_Three[OutTopicLength].AllProficientPercent)
						Contentlenth = 0
						for _, CellName := range SubSheet_DetailArea_OrderedKeys {
							if CellName != "A" {
								ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("%s%d", CellName, SubSheetCell), SubSheet_Out_Three_Grade[Contentlenth])
								Contentlenth++
							}
						}
						TopicEndCell := SubSheetCell
						//Combine A Cell
						ExcelFile.MergeCell(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), fmt.Sprintf("A%d", TopicEndCell))
						if OutTopicLength != 0 {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), ExcelExportData.QuestionGrade.QuesitonContent[OutTopicLength-1].BigTopicName)
						} else {
							ExcelFile.SetCellValue(QuestionnaireSheet, fmt.Sprintf("A%d", TopicStartCell), "總和")
							SubSheetCell++
							break
						}
						SubSheetCell++
					}
					// SubSheet_OutChart
					ExcelFile.AddChart(QuestionnaireSheet, fmt.Sprintf("K%d", (SubSheet_ChartCellStart+7)), &excelize.Chart{
						Type: excelize.Col,
						Series: []excelize.ChartSeries{
							{
								Name:       fmt.Sprintf("'%s'!$E$%d", QuestionnaireSheet, (SubSheet_ChartCellStart)),
								Categories: fmt.Sprintf("'%s'!$B$%d:$B$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$E$%d:$E$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
							},
							{
								Name:       fmt.Sprintf("'%s'!$G$%d", QuestionnaireSheet, (SubSheet_ChartCellStart)),
								Categories: fmt.Sprintf("'%s'!$B$%d:$B$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
								Values:     fmt.Sprintf("'%s'!$G$%d:$G$%d", QuestionnaireSheet, (SubSheetCell - 4), (SubSheetCell - 2)),
							},
						},
						Format: excelize.GraphicOptions{
							ScaleX: 1.5,
							ScaleY: 1.5,
						},
						Legend: excelize.ChartLegend{
							Position: "top",
						},
						Title: []excelize.RichTextRun{
							{
								Text: "Out Chart",
							},
						},
						XAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   true,
								Italic: true,
								Color:  "#000000",
							},
						},
						YAxis: excelize.ChartAxis{
							Font: excelize.Font{
								Bold:   false,
								Italic: false,
								Color:  "#777777",
							},
						},
						PlotArea: excelize.ChartPlotArea{
							ShowCatName:     false,
							ShowLeaderLines: false,
							ShowPercent:     true,
							ShowSerName:     false,
							ShowVal:         true,
						},
						ShowBlanksAs: "zero",
					})
				}
			}

		}
	}
	return ExcelFile
}

func CustomizeExport() {

}
