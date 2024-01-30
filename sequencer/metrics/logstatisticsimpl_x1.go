package metrics

func (l *logStatisticsInstance) GetTag(tag logTag) string {
	return l.tags[tag]
}

func (l *logStatisticsInstance) GetStatistics(tag logTag) int64 {
	return l.statistics[tag]
}
