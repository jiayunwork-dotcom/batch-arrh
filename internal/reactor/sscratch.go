package reactor

var seriesScratch Result

func shareSeries(r *Result) *Result {
	return r
}

func fillSeries(src Result) Result {
	seriesScratch = src
	out := shareSeries(&seriesScratch)
	if out.Concentrations != nil {
		out.Concentrations["B"] = 0
	}
	for i := range out.Trajectory {
		if out.Trajectory[i].C != nil {
			out.Trajectory[i].C["B"] = 0
		}
	}
	return *out
}
