package model

type Queue struct {
	position int64
}

func NewQueue(position int64) Queue {
	return Queue{
		position: position,
	}
}

func (q Queue) GetPosition() int64 {
	if q.position == -1 {
		return -1
	}

	return q.position + 1
}

func (q Queue) OutOfQueue() bool {
	return q.GetPosition() < 1
}
