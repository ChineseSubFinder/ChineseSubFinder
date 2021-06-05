package sub_supplier

type iSupplier interface {
	GetSubListFromFile(filePath string, httpProxy string) ([]SubInfo, error)
}