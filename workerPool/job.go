package workerpool

type STATUS uint64

const (
	WAITING = iota
	PROCESSING
	FAILED
	DONE
)

func (s STATUS) String() string {
	switch s {
	case WAITING:
		{
			return "Waiting"
		}
	case PROCESSING:
		{
			return "Processing"
		}
	case FAILED:
		{
			return "Failed"
		}
	case DONE:
		{
			return "Done"
		}
	}

	return "Undefined"
}

type Job[T any] struct {
	Id      uint64
	Element T
}

// job status
type JobStatus struct {
	Id     uint64
	Status STATUS
}
