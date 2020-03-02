package sortmode

type SortMode int

const (
	Undefined SortMode = iota
	UsageCountDesc
	UsageCountAsc
	LastUsedDesc
	LastUsedAsc
	RandomDraw
)

func ParseQuerySortMode(s string) SortMode {
	switch s {
	case "+", "＋":
		return UsageCountDesc
	case "-", "ー":
		return UsageCountAsc
	case ">", "》", "＞":
		return LastUsedDesc
	case "<", "《", "＜":
		return LastUsedAsc
	case "?", "？":
		return RandomDraw
	}
	return Undefined
}
