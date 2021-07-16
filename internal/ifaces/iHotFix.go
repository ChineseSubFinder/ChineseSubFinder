package ifaces


type IHotFix interface {

	GetKey() string

	Process() error
}