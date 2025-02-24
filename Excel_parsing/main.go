package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Student struct {
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

type ComponentRank struct {
	Emplid string
	Marks  float64
	Rank   int
}

func main() {
	f, err := excelize.OpenFile("CSF111_202425_01_GradeBook_stripped.xlsx")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		fmt.Println("No sheets found in the workbook")
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
			continue // Skip header
		}
		student := parseStudent(row)
		students = append(students, student)
	}

	computeGeneralAverages(students)
	computeBranchWiseAverages(students)
	computeRankings(students)
	computeBranchWiseRankings(students)
}

func parseStudent(row []string) Student {
	return Student{
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

func computeGeneralAverages(students []Student) {
	var sumQuiz, sumMidSem, sumLabTest, sumWeeklyLabs, sumPreCompre, sumCompre, sumTotal float64
	count := float64(len(students))

	for _, s := range students {
		sumQuiz += s.Quiz
		sumMidSem += s.MidSem
		sumLabTest += s.LabTest
		sumWeeklyLabs += s.WeeklyLabs
		sumPreCompre += s.PreCompre
		sumCompre += s.Compre
		sumTotal += s.Total
	}

	fmt.Println("General Averages:")
	fmt.Printf("Quiz: %.2f\n", sumQuiz/count)
	fmt.Printf("Mid-Sem: %.2f\n", sumMidSem/count)
	fmt.Printf("Lab Test: %.2f\n", sumLabTest/count)
	fmt.Printf("Weekly Labs: %.2f\n", sumWeeklyLabs/count)
	fmt.Printf("Pre-Compre: %.2f\n", sumPreCompre/count)
	fmt.Printf("Compre: %.2f\n", sumCompre/count)
	fmt.Printf("Total: %.2f\n\n", sumTotal/count)
}

func computeBranchWiseAverages(students []Student) {
	branchTotals := make(map[string]float64)
	branchCounts := make(map[string]int)

	for _, s := range students {
		if strings.HasPrefix(s.CampusID, "2024") {
			branch := s.CampusID[4:8]
			branchTotals[branch] += s.Total
			branchCounts[branch]++
		}
	}

	fmt.Println("Branch-wise Averages (2024 batch):")
	for branch, total := range branchTotals {
		avg := total / float64(branchCounts[branch])
		fmt.Printf("%s: %.2f\n", branch, avg)
	}
	fmt.Println()
}

func computeRankings(students []Student) {
	components := []struct {
		name  string
		score func(Student) float64
	}{
		{"Quiz", func(s Student) float64 { return s.Quiz }},
		{"Mid-Sem", func(s Student) float64 { return s.MidSem }},
		{"Lab Test", func(s Student) float64 { return s.LabTest }},
		{"Weekly Labs", func(s Student) float64 { return s.WeeklyLabs }},
		{"Pre-Compre", func(s Student) float64 { return s.PreCompre }},
		{"Compre", func(s Student) float64 { return s.Compre }},
		{"Total", func(s Student) float64 { return s.Total }},
	}

	for _, comp := range components {
		ranks := make([]ComponentRank, len(students))
		for i, s := range students {
			ranks[i] = ComponentRank{
				Emplid: s.Emplid,
				Marks:  comp.score(s),
			}
		}

		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].Marks > ranks[j].Marks
		})

		fmt.Printf("Top 3 students for %s:\n", comp.name)
		for i := 0; i < 3 && i < len(ranks); i++ {
			fmt.Printf("%s. Emplid: %s, Marks: %.2f, Rank: %s\n",
				ordinal(i+1), ranks[i].Emplid, ranks[i].Marks, ordinal(i+1))
		}
		fmt.Println()
	}
}

func computeBranchWiseRankings(students []Student) {
	branchStudents := make(map[string][]Student)

	for _, s := range students {
		if strings.HasPrefix(s.CampusID, "2024") {
			branch := s.CampusID[4:8]
			branchStudents[branch] = append(branchStudents[branch], s)
		}
	}

	fmt.Println("Branch-wise Top 3 Students (2024 batch):")
	for branch, students := range branchStudents {
		sort.Slice(students, func(i, j int) bool {
			return students[i].Total > students[j].Total
		})

		fmt.Printf("Top 3 students for branch %s:\n", branch)
		for i := 0; i < 3 && i < len(students); i++ {
			fmt.Printf("%s. Emplid: %s, Total Marks: %.2f, Rank: %s\n",
				ordinal(i+1), students[i].Emplid, students[i].Total, ordinal(i+1))
		}
		fmt.Println()
	}
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}