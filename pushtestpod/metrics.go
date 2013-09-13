package main

type Metrics map[string]int

func (m Metrics) Copy() Metrics {
	tmp := make(Metrics)
	for k, v := range m {
		tmp[k] = v
	}
	return tmp
}
