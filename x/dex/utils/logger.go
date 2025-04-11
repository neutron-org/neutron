package utils

type DataCollector struct {
	Active bool
	Data   []Entry
}

type Entry struct {
	Name string
	Data interface{}
}

func (d *DataCollector) Push(name string, data interface{}) {
	if d.Active {
		d.Data = append(d.Data, Entry{name, data})
	}
}
