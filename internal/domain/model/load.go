package model

const maxCountInLoad int64 = 1

type Load struct {
	Count int64
}

func (l *Load) ReachedCapacity() bool {
	return l.Count >= maxCountInLoad
}
