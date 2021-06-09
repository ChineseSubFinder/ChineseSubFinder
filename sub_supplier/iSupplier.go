package sub_supplier

type ISupplier interface {

	GetSupplierName() string

	GetSubListFromFile(filePath string) ([]SubInfo, error)

	GetSubListFromKeyword(keyword string) ([]SubInfo, error)
}