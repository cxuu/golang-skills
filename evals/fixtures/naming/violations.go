package util

const BAD_NAME = 1

const (
	ALSO_BAD = 2
	GoodName = 3
)

type Widget struct{}

func (this *Widget) GetName() string {
	return ""
}
