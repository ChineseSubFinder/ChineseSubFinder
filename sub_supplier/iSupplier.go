package sub_supplier

type iSupplier interface {

	GetSubListFromFile(filePath string) ([]SubInfo, error)

	GetSubListFromKeyword(keyword string) ([]SubInfo, error)
}