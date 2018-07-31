package main

type DataReader struct {
}

func (dr DataReader) Read(b []byte) (int, error) {
	for i, _ := range b {
		b[i] = byte(i)
	}
	return len(b), nil
}
