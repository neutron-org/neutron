package bindings

type BindingMarshaller interface {
	MarshalBinding() ([]byte, error)
}
