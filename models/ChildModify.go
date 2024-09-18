package models

import (
	"errors"
	"fmt"
	Routines "ginapi/routine"
	"ginapi/types"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

func UpdateChildInfo(AllInfo map[string]interface{}) string {
	var db = Routines.MeisrDB
	// 提取 BasicData
	paramData, _ := AllInfo["Param"].(map[string]interface{})
	basicData, _ := paramData["BasicData"].(map[string]interface{})
	Status, _ := paramData["Status"].(string)
	detailData, _ := paramData["DetailData"].(map[string]interface{})
	familyData, _ := paramData["FamilyData"].(map[string]interface{})
	fmt.Print(basicData)
	// 同樣的方式提取其他資料，如 DetailData, Diagnosis, Family 等
	// ...

	age, _ := basicData["Age"].(float64) // 注意這裡是 float64 類型
	birthDay, _ := basicData["BirthDay"].(string)
	t, err := time.Parse(time.RFC3339, birthDay)
	if err != nil {
		fmt.Println("Error parsing date:", err)
	}
	NewbirthDay := t.Format("2006-01-02")
	// ClassCode 和 SchoolCode 可能是數字類型，轉換為 string
	classCode := fmt.Sprintf("%v", basicData["ClassCode"])
	className, _ := basicData["ClassName"].(string)
	schoolCode := fmt.Sprintf("%v", basicData["SchoolCode"])
	schoolName, _ := basicData["SchoolName"].(string)
	semester, _ := basicData["Semester"].(string)
	studentCode := fmt.Sprintf("%v", basicData["StudentCode"]) // 可能也是數字類型
	// teacherName, _ := basicData["TeacherName"].(string)
	studentName, _ := basicData["ChildName"].(string)
	year := fmt.Sprintf("%v", basicData["Year"]) // 可能也是數字類型

	resident, _ := familyData["Resident"].(string)
	fstattend, _ := familyData["Fstattend"].(string)
	secattend, _ := familyData["Secattend"].(string)
	var proofsString string
	identites, _ := detailData["Identites"].(string)
	if proofs, ok := detailData["Proofs"].([]interface{}); ok {
		proofStrings := make([]string, len(proofs))
		for i, v := range proofs {
			if str, ok := v.(string); ok {
				proofStrings[i] = str
			}
		}
		proofsString = strings.Join(proofStrings, " ")
		detailData["Proofs"] = proofsString
	}
	degree, _ := detailData["Degree"].(string)
	diagnosis, _ := detailData["Diagnosis"].(string)
	otherdiagnosis, _ := detailData["OtherDiagnosis"].(string)
	note, _ := detailData["Note"].(string)
	manual, _ := detailData["Manual"].(string)
	placement, _ := detailData["Placement"].(string)
	fmt.Println(proofsString)
	fmt.Printf("Type: %T, Value: %#v\n", detailData["Proofs"], detailData["Proofs"])
	// diagnosis := types.ChildDiagnosis{
	// 	Identites:      getStringValue(detailData, "Identites"),
	// 	Proofs:         getStringValue(detailData, "Proofs"),
	// 	Diagnosis:      getStringValue(detailData, "Diagnosis"),
	// 	OtherDiagnosis: getStringValue(detailData, "OtherDiagnosis"),
	// 	Note:           getStringValue(detailData, "Note"),
	// 	Degree:         getStringValue(detailData, "Degree"),
	// 	Placement:      getStringValue(detailData, "Placement"),
	// 	Manual:         getStringValue(detailData, "Manual"),
	// }

	// // Mapping data to ChildFamily struct
	// family := types.ChildFamily{
	// 	Resident:      getStringValue(familyData, "Resident"),
	// 	Fstattend:     getStringValue(familyData, "Fstattend"),
	// 	Secattend:     getStringValue(familyData, "Secattend"),
	// 	OtherResident: getStringValue(familyData, "OtherResident"),
	// }
	CurrentStudentID := basicData["StudentID"].(string)
	SchoolCode_int, _ := strconv.Atoi(schoolCode)
	ClassCode_int, _ := strconv.Atoi(classCode)
	StudentCode_int, _ := strconv.Atoi(studentCode)

	if SchoolCode_int < 10 {
		schoolCode = "00" + schoolCode
	} else if SchoolCode_int < 100 {
		schoolCode = "0" + schoolCode
	}

	if ClassCode_int < 10 {
		classCode = "0" + classCode
	}
	if StudentCode_int < 10 {
		studentCode = "0" + studentCode
	}
	NewStudentID := "S" + year + schoolCode + classCode + studentCode
	fmt.Println("Current：" + CurrentStudentID)
	fmt.Println("New：" + NewStudentID)
	if NewStudentID != CurrentStudentID {
		var ChildData types.Child
		result := db.Table("studentschooltable").Select("*").Where("StudentID", NewStudentID).Take(&ChildData)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// 沒有找到匹配的記錄
				fmt.Println("No record found")
				db.Exec("UPDATE studentschooltable SET StudentID = ? WHERE StudentID = ? ", NewStudentID, CurrentStudentID)
				CurrentStudentID = NewStudentID
				fmt.Println("change id to ：" + CurrentStudentID)
			} else {
				// 其他錯誤
				fmt.Println("Error occurred:", result.Error)
			}
		}
	}
	fmt.Println(familyData)
	fmt.Println(fstattend)
	fmt.Println(secattend)
	db.Exec("UPDATE studentschooltable SET StudentName = ?, StudentCode = ?, Year = ?, Semester = ?, SchoolName = ?, SchoolCode = ?, ClassName = ?, ClassCode = ?, BirthDay = ?, Age = ?  WHERE StudentID = ? ", studentName, studentCode, year, semester, schoolName, schoolCode, className, classCode, NewbirthDay, age, CurrentStudentID)
	db.Exec("UPDATE studentstatustable SET Resident = ?, `Fst-attend` = ?, `Sec-attend` = ? WHERE StudentID = ? ", resident, fstattend, secattend, CurrentStudentID)
	if Status == "confirm" || Status == "suspected" {
		db.Exec("UPDATE studentstatustable SET Identities = ?, Proofs = ?, Manual = ?, Diagnosis = ?, OtherDiagnosis = ?,Note = ?, Degree = ?, Placement = ? WHERE StudentID = ? ", identites, proofsString, manual, diagnosis, otherdiagnosis, note, degree, placement, CurrentStudentID)
	} else if Status == "none" {
		db.Exec("UPDATE studentstatustable SET Identities = ?, Proofs = ?, Manual = ?, Diagnosis = ?, OtherDiagnosis = ?,Note = ?, Degree = ?, Placement = ? WHERE StudentID = ? ", nil, nil, nil, nil, nil, nil, nil, nil, CurrentStudentID)
	}
	return "update successful"
}
