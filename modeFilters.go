package main

type ModeFilter []int

var (
	BusFilter = ModeFilter{
		1,
		3,
		15,
		4,
		12,
	}
	MetroFilter = ModeFilter{0}
	TrainFilter = ModeFilter{2}
	NoFilter    = NewFilter(BusFilter, MetroFilter, TrainFilter)
)

func NewFilter(filters ...ModeFilter) ModeFilter {
	retFilter := ModeFilter{}
	for _, v := range filters {
		retFilter = append(retFilter, v...)
	}
	return retFilter
}
