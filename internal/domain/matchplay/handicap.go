package matchplay

func strokesReceived(handicap float64, strokeIndex int, totalHoles int) int {
	if totalHoles <= 0 {
		return 0
	}
	h := int(handicap + 0.5)
	if h < 0 {
		h = 0
	}
	base := h / totalHoles
	remainder := h % totalHoles
	strokes := base
	if strokeIndex <= remainder {
		strokes++
	}
	return strokes
}
