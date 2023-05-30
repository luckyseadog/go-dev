package internal

type Queue struct {
	Elements []Counter
	Size     Counter
}

func NewQueue(size Counter) *Queue {
	return &Queue{Elements: []Counter{}, Size: size}
}

func (q *Queue) IsEmpty() bool {
	return len(q.Elements) == 0
}
func (q *Queue) GetLength() Counter {
	return Counter(len(q.Elements))
}

func (q *Queue) Push(elem Counter) {
	q.Elements = append(q.Elements, elem)
}

func (q *Queue) Pop() Counter {
	element := q.Elements[0]
	if q.GetLength() == 1 {
		q.Elements = nil
		return element
	}
	q.Elements = q.Elements[1:]
	return element
}