package ifaces

type IHotFix interface {
	GetKey() string

	Process() (interface{}, error)
}
