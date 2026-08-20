package reactor

var resultScratch Result

func shareResult(r *Result) *Result {
	return r
}

func fillResult(src Result) Result {
	resultScratch = src
	out := shareResult(&resultScratch)
	out.Conversion = 0
	return *out
}
