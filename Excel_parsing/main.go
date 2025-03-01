package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type Student struct {
	ClassNo    string
	Emplid     string
	CampusID   string
	Quiz       float64
	MidSem     float64
	LabTest    float64
	WeeklyLabs float64
	PreCompre  float64
	Compre     float64
	Total      float64
}

type SummaryReport struct {
	GeneralAverages   map[string]float64 `json:"general_averages"`
	BranchAverages    map[string]float64 `json:"branch_averages"`
	BranchRankings    map[string][]Student `json:"branch_rankings"`
	OverallTopStudents []Student `json:"overall_top_students"`
}

func main() {
	exportFormat := flag.String("export", "", "Export final report to json format")
	classFilter := flag.String("class", "", "Filter by input Class No.")
	flag.Parse()

	f, err := excelize.OpenFile("CSF111_202425_01_GradeBook_stripped.xlsx")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		fmt.Println("No sheets found in the file")
		return
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println("Error reading sheet:", err)
		return
	}

	var students []Student
	for i, row := range rows {
		if i == 0 {
			continue // Skipped 1st line
		}
		student := parseStudent(row)
		if *classFilter == "" || student.ClassNo == *classFilter {
			students = append(students, student)
		}
	}

	report := generateReport(students)

	if *exportFormat == "json" {
		exportToJson(report)
	} else {
		printReport(report)
	}
}

func parseStudent(row []string) Student {
	return Student{
		ClassNo:    row[1],
		Emplid:     row[2],
		CampusID:   row[3],
		Quiz:       parseFloat(row[4]),
		MidSem:     parseFloat(row[5]),
		LabTest:    parseFloat(row[6]),
		WeeklyLabs: parseFloat(row[7]),
		PreCompre:  parseFloat(row[8]),
		Compre:     parseFloat(row[9]),
		Total:      parseFloat(row[10]),
	}
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func generateReport(students []Student) SummaryReport {
	generalAverages := GeneralAverages(students)
	branchAverages := BranchWiseAverages(students)
	branchRankings := BranchWiseRankings(students)
	overallTopStudents := OverallTopStudents(students)

	return SummaryReport{
		GeneralAverages:    generalAverages,
		BranchAverages:     branchAverages,
		BranchRankings:     branchRankings,
		OverallTopStudents: overallTopStudents,
	}
}

func OverallTopStudents(students []Student) []Student {
	sort.Slice(students, func(i, j int) bool {
		return students[i].Total > students[j].Total
	})
	if len(students) > 3 {
		return students[:3]
	}
	return students
}

func GeneralAverages(students []Student) map[string]float64 {
	averages := make(map[string]float64)
	var sumTotal float64
	count := float64(len(students))

	for _, s := range students {
		sumTotal += s.Total
	}
	averages["Total"] = sumTotal / count
	return averages
}

func BranchWiseAverages(students []Student) map[string]float64 {
	branchAverages := make(map[string]float64)
	branchCounts := make(map[string]int)

	for _, s := range students {
		branch := s.CampusID[4:8]
		branchAverages[branch] += s.Total
		branchCounts[branch]++
	}

	for branch, total := range branchAverages {
		branchAverages[branch] = total / float64(branchCounts[branch])
	}
	return branchAverages
}

func BranchWiseRankings(students []Student) map[string][]Student {
	branchRankings := make(map[string][]Student)
	branchStudents := make(map[string][]Student)

	for _, s := range students {
		branch := s.CampusID[4:8]
		branchStudents[branch] = append(branchStudents[branch], s)
	}

	for branch, studs := range branchStudents {
		sort.Slice(studs, func(i, j int) bool {
			return studs[i].Total > studs[j].Total
		})
		branchRankings[branch] = studs
	}
	return branchRankings
}

func exportToJson(report SummaryReport) {
	file, err := os.Create("summary_report.json")
	if err != nil {
		fmt.Println("Error creating JSON file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Println("Error writing to JSON file:", err)
	} else {
		fmt.Println("Summary report successfully exported to summary_report.json")
	}
}

func printReport(report SummaryReport) {
	fmt.Println("General Averages:", report.GeneralAverages)
	fmt.Println("Branch Averages:", report.BranchAverages)
	fmt.Println("Branch Rankings:", report.BranchRankings)
	fmt.Println("Overall Top Students:", report.OverallTopStudents)
}
