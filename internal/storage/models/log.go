package models

type Log struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Line      *int64 `json:"line,omitempty"`
	Content   string `json:"content,omitempty"`
}

type Logs []*Log

func (l Logs) Len() int {
	return len(l)
}

func (l Logs) Less(i, j int) bool {
	if l[i].Line == nil || l[j].Line == nil {
		return l[i].Timestamp < l[j].Timestamp
	}
	return *l[i].Line < *l[j].Line
}

func (l Logs) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
