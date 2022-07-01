package backend

type CheckSubSupplier struct {
	SupplierNames []string `json:"supplier_names" binding:"required"`
}
