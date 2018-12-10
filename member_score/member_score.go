package member_score

import "strings"

type MemberScore struct {
	Id int
	Day int
	Name string
	Part1 int64
	Part2 int64
	AocLocalScore int
	AocGlobalScore int
	Count int64
}

func (m MemberScore) Part1Avg() int64 {
	return m.Part1 / m.Count
}

func (m MemberScore) Part2Avg() int64 {
	if m.Part2 == 0 {
		return 0
	}
	return m.Part2 / m.Count
}

func (m MemberScore) Part2DiffAvg() int64 {
	if m.Part2Diff() < 1 {
		return 0
	}
	return m.Part2Diff() / m.Count
}

func (m MemberScore) Part2Diff() int64 {
	if m.Part2 == 0 {
		return -1
	}
	return m.Part2 - m.Part1
}

type ByName []*MemberScore
func (a ByName) Len() int { return len(a) }
func (a ByName) Less(i, j int) bool { return strings.ToLower(a[i].Name) < strings.ToLower(a[j].Name) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPart1 []*MemberScore
func (a ByPart1) Len() int { return len(a) }
func (a ByPart1) Less(i, j int) bool { return a[i].Part1 < a[j].Part1 }
func (a ByPart1) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPart2 []*MemberScore
func (a ByPart2) Len() int { return len(a) }
func (a ByPart2) Less(i, j int) bool {
	if a[i].Part2 == 0 && a[j].Part2 > 0 {
		return false
	}
	if a[j].Part2 == 0 && a[i].Part2 > 0 {
		return true
	}
	return a[i].Part2 < a[j].Part2
}
func (a ByPart2) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByAocLocalScore []*MemberScore
func (a ByAocLocalScore) Len() int { return len(a) }
func (a ByAocLocalScore) Less(i, j int) bool { return a[i].AocLocalScore > a[j].AocLocalScore }
func (a ByAocLocalScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByAocGlobalScore []*MemberScore
func (a ByAocGlobalScore) Len() int { return len(a) }
func (a ByAocGlobalScore) Less(i, j int) bool {
	if a[i].AocGlobalScore == a[j].AocGlobalScore {
		return a[i].AocLocalScore > a[j].AocLocalScore
	}

	return a[i].AocGlobalScore > a[j].AocGlobalScore
}
func (a ByAocGlobalScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPart2Diff []*MemberScore
func (a ByPart2Diff) Len() int { return len(a) }
func (a ByPart2Diff) Less(i, j int) bool {
	if a[i].Count != a[j].Count {
		return a[i].Count > a[j].Count
	}

	if a[i].Part2Diff() == -1 && a[j].Part2Diff() > -1 {
		return false
	}
	if a[i].Part2Diff() > -1 && a[j].Part2Diff() == -1 {
		return true
	}

	if a[i].Part2Diff() == a[j].Part2Diff() {
		return a[i].Part1 < a[j].Part1
	}

	return a[i].Part2Diff() < a[j].Part2Diff()
}
func (a ByPart2Diff) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

